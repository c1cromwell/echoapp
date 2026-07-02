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

const defaultEstuaryAPIBase = "https://api.estuary.tech"

// FilecoinArchiver creates long-term Filecoin deals via Estuary (WO-185 subset).
type FilecoinArchiver struct {
	apiBase string
	token   string
	client  *http.Client
}

// NewFilecoinArchiver reads ESTUARY_API_TOKEN and optional ESTUARY_API_BASE.
func NewFilecoinArchiver() (*FilecoinArchiver, error) {
	token := os.Getenv("ESTUARY_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("filecoin: ESTUARY_API_TOKEN not configured")
	}
	base := os.Getenv("ESTUARY_API_BASE")
	if base == "" {
		base = defaultEstuaryAPIBase
	}
	return &FilecoinArchiver{
		apiBase: strings.TrimRight(base, "/"),
		token:   token,
		client:  &http.Client{Timeout: 90 * time.Second},
	}, nil
}

// FilecoinDeal describes a storage deal request outcome.
type FilecoinDeal struct {
	CID      string `json:"cid"`
	DealID   string `json:"deal_id,omitempty"`
	Provider string `json:"provider,omitempty"`
	Status   string `json:"status"`
}

// CreateDeal pins an existing IPFS CID for Filecoin replication.
func (f *FilecoinArchiver) CreateDeal(ctx context.Context, cid string, durationDays int) (*FilecoinDeal, error) {
	if cid == "" {
		return nil, fmt.Errorf("filecoin: empty cid")
	}
	if durationDays <= 0 {
		durationDays = 180
	}
	payload := map[string]interface{}{
		"cid":      cid,
		"duration": durationDays,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.apiBase+"/pinning/pins", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.token)
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("filecoin deal: %s", string(b))
	}
	var out struct {
		Pin struct {
			CID    string `json:"cid"`
			DealID int64  `json:"deal_id"`
		} `json:"pin"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return &FilecoinDeal{CID: cid, Status: "submitted"}, nil
	}
	return &FilecoinDeal{
		CID:    out.Pin.CID,
		DealID: fmt.Sprintf("%d", out.Pin.DealID),
		Status: "submitted",
	}, nil
}

// ArchivingBackend optionally archives CIDs to Filecoin after IPFS pin.
type ArchivingBackend struct {
	StorageBackend
	archiver *FilecoinArchiver
	mu       sync.Mutex
	deals    map[string]*FilecoinDeal
}

// NewArchivingBackend wraps a content backend with optional Filecoin deals.
func NewArchivingBackend(storage StorageBackend, archiver *FilecoinArchiver) *ArchivingBackend {
	return &ArchivingBackend{
		StorageBackend: storage,
		archiver:       archiver,
		deals:          make(map[string]*FilecoinDeal),
	}
}

// Store pins data and optionally creates a Filecoin deal when configured.
func (a *ArchivingBackend) Store(ctx context.Context, key string, data []byte) (string, error) {
	cid, err := a.StorageBackend.Store(ctx, key, data)
	if err != nil || cid == "" || a.archiver == nil {
		return cid, err
	}
	deal, derr := a.archiver.CreateDeal(ctx, cid, 180)
	if derr == nil && deal != nil {
		a.mu.Lock()
		a.deals[key] = deal
		a.mu.Unlock()
	}
	return cid, err
}

// DealForKey returns the last Filecoin deal for a storage key, if any.
func (a *ArchivingBackend) DealForKey(key string) *FilecoinDeal {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.deals[key]
}
