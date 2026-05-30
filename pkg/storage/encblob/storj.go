package encblob

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// StorjStorage pushes encrypted blobs to Storj's S3-compatible gateway.
type StorjStorage struct {
	accessKey string
	secretKey string
	bucket    string
	endpoint  string
	client    *http.Client
}

// NewStorjStorage reads credentials from env vars.
func NewStorjStorage() (*StorjStorage, error) {
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
	return &StorjStorage{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		endpoint:  endpoint,
		client:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Store uploads ciphertext and returns a content-addressed URI.
func (s *StorjStorage) Store(ctx context.Context, encrypted []byte) (string, error) {
	h := sha256.Sum256(encrypted)
	objectKey := fmt.Sprintf("passport-sync/%x", h)

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
