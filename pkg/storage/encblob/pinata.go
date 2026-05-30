package encblob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const pinataAPIBase = "https://api.pinata.cloud"

// PinataStorage pushes encrypted blobs to Pinata's IPFS pinning service.
type PinataStorage struct {
	apiKey    string
	apiSecret string
	client    *http.Client
}

// NewPinataStorage reads credentials from PINATA_API_KEY / PINATA_API_SECRET env vars.
func NewPinataStorage() (*PinataStorage, error) {
	key := os.Getenv("PINATA_API_KEY")
	secret := os.Getenv("PINATA_API_SECRET")
	if key == "" || secret == "" {
		return nil, ErrStorageNotConfigured
	}
	return &PinataStorage{
		apiKey:    key,
		apiSecret: secret,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Store pins ciphertext to IPFS via Pinata and returns the CID.
func (p *PinataStorage) Store(ctx context.Context, encrypted []byte) (string, error) {
	payload := map[string]interface{}{
		"pinataContent": map[string]interface{}{
			"data":      encrypted,
			"algorithm": "AES-256-GCM",
		},
		"pinataOptions": map[string]interface{}{
			"cidVersion": 1,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("pinata marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		pinataAPIBase+"/pinning/pinJSONToIPFS", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("pinata_api_key", p.apiKey)
	req.Header.Set("pinata_secret_api_key", p.apiSecret)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("pinata request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pinata returned %d: %s", resp.StatusCode, b)
	}

	var result struct {
		IpfsHash string `json:"IpfsHash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("pinata decode response: %w", err)
	}
	if result.IpfsHash == "" {
		return "", fmt.Errorf("pinata returned empty CID")
	}
	return result.IpfsHash, nil
}
