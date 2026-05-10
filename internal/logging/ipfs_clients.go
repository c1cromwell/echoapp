package logging

// WO-33: Decentralized storage integration for encrypted audit log batches.
//
// Two concrete IPFSStorage implementations:
//   - PinataIPFSStorage — Pinata v1 pinning API (primary)
//   - StorjIPFSStorage  — Storj S3-compatible API (fallback)
//
// FallbackIPFSStorage composes them: try Pinata first; on failure fall back to
// Storj; fire an async secondary pin for redundancy.
//
// Configuration (set in .env.local):
//   PINATA_API_KEY / PINATA_API_SECRET
//   STORJ_ACCESS_KEY / STORJ_SECRET_KEY / STORJ_BUCKET / STORJ_ENDPOINT

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ErrStorageNotConfigured is returned when required environment credentials are absent.
var ErrStorageNotConfigured = errors.New("storage provider not configured")

// --- Pinata IPFS client ---

const pinataAPIBase = "https://api.pinata.cloud"

// PinataIPFSStorage pushes encrypted log batches to Pinata's IPFS pinning service.
type PinataIPFSStorage struct {
	apiKey    string
	apiSecret string
	client    *http.Client
}

// NewPinataIPFSStorage reads credentials from PINATA_API_KEY / PINATA_API_SECRET env vars.
func NewPinataIPFSStorage() (*PinataIPFSStorage, error) {
	key := os.Getenv("PINATA_API_KEY")
	secret := os.Getenv("PINATA_API_SECRET")
	if key == "" || secret == "" {
		return nil, ErrStorageNotConfigured
	}
	return &PinataIPFSStorage{
		apiKey:    key,
		apiSecret: secret,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Store pins the encrypted batch to IPFS via Pinata and returns the CID.
func (p *PinataIPFSStorage) Store(ctx context.Context, encrypted []byte) (string, error) {
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

// --- Storj S3-compatible client ---

// StorjIPFSStorage pushes encrypted batches to Storj's S3-compatible gateway.
type StorjIPFSStorage struct {
	accessKey string
	secretKey string
	bucket    string
	endpoint  string
	client    *http.Client
}

// NewStorjIPFSStorage reads credentials from env vars.
func NewStorjIPFSStorage() (*StorjIPFSStorage, error) {
	accessKey := os.Getenv("STORJ_ACCESS_KEY")
	secretKey := os.Getenv("STORJ_SECRET_KEY")
	bucket := os.Getenv("STORJ_BUCKET")
	endpoint := os.Getenv("STORJ_ENDPOINT")
	if accessKey == "" || secretKey == "" || bucket == "" {
		return nil, ErrStorageNotConfigured
	}
	if endpoint == "" {
		endpoint = "https://gateway.storjshare.io"
	}
	return &StorjIPFSStorage{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		endpoint:  endpoint,
		client:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Store uploads the encrypted batch to Storj and returns a content-addressed key.
// The object key is SHA-256(ciphertext) hex, giving a deterministic CID-like ID.
//
// Note: production Storj requires AWS Signature V4; this scaffold uses Basic Auth
// which is accepted by the Storj S3 gateway for development. Wire proper signing
// via aws/aws-sdk-go-v2 or minio-go before production deployment.
func (s *StorjIPFSStorage) Store(ctx context.Context, encrypted []byte) (string, error) {
	h := sha256.Sum256(encrypted)
	objectKey := fmt.Sprintf("audit-logs/%x", h)

	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(encrypted))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-amz-content-sha256", fmt.Sprintf("%x", h))
	req.SetBasicAuth(s.accessKey, s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("storj request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("storj returned %d: %s", resp.StatusCode, b)
	}

	return fmt.Sprintf("storj://%s/%s", s.bucket, objectKey), nil
}

// --- FallbackIPFSStorage (primary + secondary pin) ---

// FallbackIPFSStorage tries primary first; on failure uses secondary.
// On primary success it fires an async secondary pin for redundancy.
type FallbackIPFSStorage struct {
	primary   IPFSStorage
	secondary IPFSStorage
}

// NewFallbackIPFSStorage builds the production two-provider storage from env vars.
// Returns ErrStorageNotConfigured only when both providers are unavailable.
func NewFallbackIPFSStorage() (*FallbackIPFSStorage, error) {
	pinata, err1 := NewPinataIPFSStorage()
	storj, err2 := NewStorjIPFSStorage()
	switch {
	case err1 != nil && err2 != nil:
		return nil, ErrStorageNotConfigured
	case err1 != nil:
		return &FallbackIPFSStorage{primary: storj, secondary: &StubIPFSStorage{}}, nil
	case err2 != nil:
		return &FallbackIPFSStorage{primary: pinata, secondary: &StubIPFSStorage{}}, nil
	default:
		return &FallbackIPFSStorage{primary: pinata, secondary: storj}, nil
	}
}

// Store tries primary; falls back to secondary on error.
// On primary success, fires an async secondary pin for redundancy (non-blocking).
func (f *FallbackIPFSStorage) Store(ctx context.Context, encrypted []byte) (string, error) {
	cid, err := f.primary.Store(ctx, encrypted)
	if err == nil {
		encCopy := make([]byte, len(encrypted))
		copy(encCopy, encrypted)
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, _ = f.secondary.Store(bgCtx, encCopy)
		}()
		return cid, nil
	}

	// Primary failed — use secondary directly.
	cid, fallbackErr := f.secondary.Store(ctx, encrypted)
	if fallbackErr != nil {
		return "", fmt.Errorf("all storage providers failed: primary=%v fallback=%v", err, fallbackErr)
	}
	return cid, nil
}
