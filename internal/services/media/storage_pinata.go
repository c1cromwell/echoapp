package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const defaultPinataAPIBase = "https://api.pinata.cloud"

// PinataStorage pins encrypted media chunks to Pinata IPFS (WO-21 / WO-185 subset).
type PinataStorage struct {
	apiBase   string
	apiKey    string
	apiSecret string
	client    *http.Client
	gateway   string
	mu        sync.RWMutex
	keyToCID  map[string]string
}

// NewPinataStorage reads PINATA_API_KEY / PINATA_API_SECRET.
func NewPinataStorage() (*PinataStorage, error) {
	key := os.Getenv("PINATA_API_KEY")
	secret := os.Getenv("PINATA_API_SECRET")
	if key == "" || secret == "" {
		return nil, fmt.Errorf("pinata: credentials not configured")
	}
	gateway := os.Getenv("PINATA_GATEWAY")
	if gateway == "" {
		gateway = "https://gateway.pinata.cloud"
	}
	apiBase := os.Getenv("PINATA_API_BASE")
	if apiBase == "" {
		apiBase = defaultPinataAPIBase
	}
	return &PinataStorage{
		apiBase:   strings.TrimRight(apiBase, "/"),
		apiKey:    key,
		apiSecret: secret,
		client:    &http.Client{Timeout: 60 * time.Second},
		gateway:   gateway,
		keyToCID:  make(map[string]string),
	}, nil
}

// Store pins data and maps caller key → CID for retrieval.
func (p *PinataStorage) Store(ctx context.Context, key string, data []byte) (string, error) {
	if key == "" {
		return "", fmt.Errorf("pinata: empty key")
	}
	payload := map[string]interface{}{
		"pinataContent": data,
		"pinataMetadata": map[string]string{
			"name":    "echo-media-" + key,
			"echoKey": key,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiBase+"/pinning/pinJSONToIPFS", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("pinata_api_key", p.apiKey)
	req.Header.Set("pinata_secret_api_key", p.apiSecret)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pinata store: %s", string(b))
	}
	var out struct {
		IpfsHash string `json:"IpfsHash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	p.mu.Lock()
	p.keyToCID[key] = out.IpfsHash
	p.mu.Unlock()
	return out.IpfsHash, nil
}

// Retrieve fetches pinned content by key via cached CID.
func (p *PinataStorage) Retrieve(ctx context.Context, key string) ([]byte, error) {
	p.mu.RLock()
	cid, ok := p.keyToCID[key]
	p.mu.RUnlock()
	if !ok || cid == "" {
		return nil, fmt.Errorf("pinata: unknown key %q", key)
	}
	gateway := p.gateway
	url := strings.TrimRight(gateway, "/") + "/ipfs/" + cid
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pinata retrieve: status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// Delete is a no-op for Pinata MVP (unpin deferred).
func (p *PinataStorage) Delete(_ context.Context, _ string) error { return nil }
