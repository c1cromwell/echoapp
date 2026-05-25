package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// IPFSStorage is a content-addressed StorageBackend backed by an IPFS (kubo) node
// via its HTTP API, using the Mutable File System (MFS) to map caller keys to
// stable paths. Bytes are stored content-addressed in IPFS (the node computes the
// CID); MFS provides the key→CID naming so the existing key-addressed interface is
// satisfied without a separate, centralized key→CID index.
//
// This replaces S3/MinIO for media when STORAGE_BACKEND=ipfs, removing the
// centralized object-store dependency (decentralization gap D3).
type IPFSStorage struct {
	apiURL string
	root   string
	client *http.Client
}

// IPFSConfig configures the IPFS backend.
type IPFSConfig struct {
	APIURL string // kubo HTTP API base, e.g. http://localhost:5001
	Root   string // MFS root for Echo media; defaults to /echo
}

// NewIPFSStorage builds an IPFS-backed store. It does not dial the node here;
// connectivity errors surface on first use.
func NewIPFSStorage(cfg IPFSConfig) (*IPFSStorage, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("ipfs: APIURL is required")
	}
	root := cfg.Root
	if root == "" {
		root = "/echo"
	}
	return &IPFSStorage{
		apiURL: strings.TrimRight(cfg.APIURL, "/"),
		root:   "/" + strings.Trim(root, "/"),
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// mfsPath maps a caller key to an MFS path, rejecting traversal so a key cannot
// escape the Echo media root within the MFS namespace.
func (s *IPFSStorage) mfsPath(key string) (string, error) {
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return "", fmt.Errorf("ipfs: empty key")
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == ".." || seg == "." || seg == "" {
			return "", fmt.Errorf("ipfs: invalid key %q", key)
		}
	}
	return s.root + "/" + key, nil
}

// Store writes data to MFS at the key's path (content-addressed in IPFS) and
// returns the resulting CID so callers can anchor it on-chain (D3).
func (s *IPFSStorage) Store(ctx context.Context, key string, data []byte) (string, error) {
	path, err := s.mfsPath(key)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("data", strings.TrimPrefix(path, "/"))
	if err != nil {
		return "", fmt.Errorf("ipfs: form file: %w", err)
	}
	if _, err := fw.Write(data); err != nil {
		return "", fmt.Errorf("ipfs: write part: %w", err)
	}
	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("ipfs: close multipart: %w", err)
	}

	q := url.Values{}
	q.Set("arg", path)
	q.Set("create", "true")
	q.Set("parents", "true")
	q.Set("truncate", "true")
	resp, err := s.post(ctx, "/api/v0/files/write?"+q.Encode(), mw.FormDataContentType(), &body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	return s.statCID(ctx, path)
}

// statCID returns the IPFS CID (content hash) of an MFS path via files/stat.
func (s *IPFSStorage) statCID(ctx context.Context, path string) (string, error) {
	q := url.Values{}
	q.Set("arg", path)
	resp, err := s.post(ctx, "/api/v0/files/stat?"+q.Encode(), "", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var stat struct {
		Hash string `json:"Hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stat); err != nil {
		return "", fmt.Errorf("ipfs: decode files/stat: %w", err)
	}
	return stat.Hash, nil
}

// Retrieve reads the bytes stored at the key's MFS path.
func (s *IPFSStorage) Retrieve(ctx context.Context, key string) ([]byte, error) {
	path, err := s.mfsPath(key)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("arg", path)
	resp, err := s.post(ctx, "/api/v0/files/read?"+q.Encode(), "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Delete removes the key's MFS path.
func (s *IPFSStorage) Delete(ctx context.Context, key string) error {
	path, err := s.mfsPath(key)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("arg", path)
	q.Set("force", "true")
	resp, err := s.post(ctx, "/api/v0/files/rm?"+q.Encode(), "", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// post issues a POST to the kubo API and returns the response on 2xx, or an error
// (with the response body) otherwise. The caller must close resp.Body on success.
func (s *IPFSStorage) post(ctx context.Context, path, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+path, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ipfs: request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, fmt.Errorf("ipfs: %s -> %d: %s", path, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return resp, nil
}
