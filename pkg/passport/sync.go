package passport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

var (
	ErrSyncNotFound     = errors.New("sync blob not found")
	ErrSyncHashMismatch = errors.New("content_hash does not match ciphertext")
)

// SyncBlob is metadata + opaque ciphertext for holder credential wallet sync (T2).
type SyncBlob struct {
	HolderDID   string    `json:"holder_did"`
	StorageURI  string    `json:"storage_uri"`
	ContentHash string    `json:"content_hash"`
	ByteSize    int       `json:"byte_size"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SyncBlobRecord includes ciphertext for pull (client decrypts locally).
type SyncBlobRecord struct {
	SyncBlob
	Ciphertext []byte `json:"-"`
}

// PushSyncRequest is the POST /v1/passport/sync body (base64 ciphertext from client).
type PushSyncRequest struct {
	CiphertextBase64 string `json:"ciphertext_base64"`
	ContentHash      string `json:"content_hash,omitempty"`
}

// SyncStore persists sync metadata and ciphertext relay.
type SyncStore interface {
	UpsertSyncBlob(ctx context.Context, record SyncBlobRecord) error
	GetSyncBlob(ctx context.Context, holderDID string) (*SyncBlobRecord, error)
}

// SyncService pushes/pulls client-encrypted credential bundles.
type SyncService struct {
	store   SyncStore
	storage encblob.Storage
}

func NewSyncService(store SyncStore, storage encblob.Storage) *SyncService {
	return &SyncService{store: store, storage: storage}
}

func (s *SyncService) Push(ctx context.Context, holderDID string, req PushSyncRequest) (*SyncBlob, error) {
	ciphertext, err := decodeBase64(req.CiphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext_base64: %w", err)
	}
	if len(ciphertext) == 0 {
		return nil, errors.New("ciphertext is required")
	}

	hash := strings.ToLower(strings.TrimSpace(req.ContentHash))
	if hash == "" {
		sum := sha256.Sum256(ciphertext)
		hash = hex.EncodeToString(sum[:])
	}
	if len(hash) != 64 {
		return nil, errors.New("content_hash must be 64-char hex SHA-256")
	}
	sum := sha256.Sum256(ciphertext)
	if hex.EncodeToString(sum[:]) != hash {
		return nil, ErrSyncHashMismatch
	}

	uri, err := s.storage.Store(ctx, ciphertext)
	if err != nil {
		return nil, err
	}

	existing, _ := s.store.GetSyncBlob(ctx, holderDID)
	version := 1
	if existing != nil {
		version = existing.Version + 1
	}
	now := time.Now().UTC()
	record := SyncBlobRecord{
		SyncBlob: SyncBlob{
			HolderDID:   holderDID,
			StorageURI:  uri,
			ContentHash: hash,
			ByteSize:    len(ciphertext),
			Version:     version,
			UpdatedAt:   now,
		},
		Ciphertext: ciphertext,
	}
	if err := s.store.UpsertSyncBlob(ctx, record); err != nil {
		return nil, err
	}
	return &record.SyncBlob, nil
}

func (s *SyncService) Pull(ctx context.Context, holderDID string) (*SyncBlobRecord, error) {
	record, err := s.store.GetSyncBlob(ctx, holderDID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrSyncNotFound
	}
	return record, nil
}

func decodeBase64(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty input")
	}
	// encoding/base64 import kept local to avoid unused in other files — use stdlib
	return decodeStdBase64(raw)
}
