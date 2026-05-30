package encblob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
)

// ErrStorageNotConfigured is returned when required environment credentials are absent.
var ErrStorageNotConfigured = errors.New("storage provider not configured")

// Storage pins opaque ciphertext blobs (client-encrypted; server holds no keys).
type Storage interface {
	Store(ctx context.Context, encrypted []byte) (uri string, err error)
}

// RetrievableStorage can pull ciphertext back by URI (dev stub; production clients may fetch IPFS directly).
type RetrievableStorage interface {
	Storage
	Retrieve(ctx context.Context, uri string) ([]byte, error)
}

// StubStorage records blobs in memory — used in tests and when Pinata/Storj are not configured.
type StubStorage struct {
	mu    sync.Mutex
	blobs map[string][]byte
}

func NewStubStorage() *StubStorage {
	return &StubStorage{blobs: make(map[string][]byte)}
}

func (s *StubStorage) Store(_ context.Context, encrypted []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blobs == nil {
		s.blobs = make(map[string][]byte)
	}
	h := sha256.Sum256(encrypted)
	uri := "stub-cid-" + hex.EncodeToString(h[:8])
	s.blobs[uri] = append([]byte(nil), encrypted...)
	return uri, nil
}

func (s *StubStorage) Retrieve(_ context.Context, uri string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blobs == nil {
		return nil, errors.New("blob not found")
	}
	b, ok := s.blobs[uri]
	if !ok {
		return nil, errors.New("blob not found")
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out, nil
}

// StoredURIs returns URIs produced so far (for test assertions).
func (s *StubStorage) StoredURIs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.blobs))
	for uri := range s.blobs {
		out = append(out, uri)
	}
	return out
}

// StoredCIDs is an alias for StoredURIs (audit-log tests).
func (s *StubStorage) StoredCIDs() []string { return s.StoredURIs() }
