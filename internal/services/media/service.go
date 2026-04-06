// Package media implements the encrypted media upload/download service.
package media

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
)

const (
	ChunkSize = 256 * 1024 // 256KB chunks
)

// Trust tier max file sizes.
var MaxSizeForTier = map[int]int64{
	1: 0,                      // Tier 1: no media
	2: 10 * 1024 * 1024,       // 10MB
	3: 100 * 1024 * 1024,      // 100MB
	4: 500 * 1024 * 1024,      // 500MB
	5: 2 * 1024 * 1024 * 1024, // 2GB
}

// AllowedContentTypes lists permitted media types.
var AllowedContentTypes = map[string]bool{
	"image/jpeg":               true,
	"image/png":                true,
	"image/gif":                true,
	"image/webp":               true,
	"video/mp4":                true,
	"video/quicktime":          true,
	"audio/aac":                true,
	"audio/mp4":                true,
	"application/pdf":          true,
	"application/octet-stream": true,
}

var (
	ErrTierRestricted = errors.New("trust tier does not permit media uploads")
	ErrFileTooLarge   = errors.New("file exceeds maximum size for trust tier")
	ErrInvalidContent = errors.New("content type not allowed")
	ErrFileNotFound   = errors.New("file not found")
	ErrFileFlagged    = errors.New("file flagged by content scan")
)

// UploadRequest contains upload parameters.
type UploadRequest struct {
	UploaderDID   string `json:"uploaderDid"`
	ContentType   string `json:"contentType"`
	EncryptedSize int64  `json:"encryptedSize"`
	TrustTier     int    `json:"trustTier"`
}

// UploadResult contains the result of an upload.
type UploadResult struct {
	FileID      string    `json:"fileId"`
	ChunkCount  int       `json:"chunkCount"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	Timestamp   time.Time `json:"timestamp"`
}

// StorageBackend abstracts the underlying storage (Storj/S3/IPFS).
type StorageBackend interface {
	Store(ctx context.Context, key string, data []byte) error
	Retrieve(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// Service provides media upload/download operations.
type Service struct {
	db      database.DB
	storage StorageBackend
}

// NewService creates a media service.
func NewService(db database.DB, storage StorageBackend) *Service {
	return &Service{db: db, storage: storage}
}

// Upload processes an encrypted file upload.
func (s *Service) Upload(ctx context.Context, req UploadRequest, body io.Reader) (*UploadResult, error) {
	// Validate tier
	maxSize, ok := MaxSizeForTier[req.TrustTier]
	if !ok || maxSize == 0 {
		return nil, ErrTierRestricted
	}

	// Validate size
	if req.EncryptedSize > maxSize {
		return nil, ErrFileTooLarge
	}

	// Validate content type
	if !AllowedContentTypes[req.ContentType] {
		return nil, ErrInvalidContent
	}

	fileID := uuid.New().String()

	// Read and chunk the data
	chunkIndex := 0
	buf := make([]byte, ChunkSize)
	var totalSize int64

	for {
		n, err := body.Read(buf)
		if n > 0 {
			chunkData := make([]byte, n)
			copy(chunkData, buf[:n])

			chunkID := fmt.Sprintf("%s-chunk-%d", fileID, chunkIndex)
			checksum := fmt.Sprintf("%x", sha256.Sum256(chunkData))

			// Store chunk data
			if s.storage != nil {
				if err := s.storage.Store(ctx, chunkID, chunkData); err != nil {
					return nil, err
				}
			}

			// Store chunk metadata
			chunk := &database.MediaChunk{
				ChunkID:  chunkID,
				FileID:   fileID,
				Index:    chunkIndex,
				Size:     int64(n),
				Checksum: checksum,
			}
			if err := s.db.StoreChunk(ctx, chunk); err != nil {
				return nil, err
			}

			totalSize += int64(n)
			chunkIndex++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	// Store file metadata
	file := &database.MediaFile{
		FileID:        fileID,
		UploaderDID:   req.UploaderDID,
		ContentType:   req.ContentType,
		EncryptedSize: totalSize,
		ChunkCount:    chunkIndex,
		ScanStatus:    "pending",
	}
	if err := s.db.StoreMediaFile(ctx, file); err != nil {
		return nil, err
	}

	return &UploadResult{
		FileID:      fileID,
		ChunkCount:  chunkIndex,
		ContentType: req.ContentType,
		Size:        totalSize,
		Timestamp:   time.Now(),
	}, nil
}

// Download retrieves file metadata for download.
func (s *Service) Download(ctx context.Context, fileID, requesterDID string) ([]byte, *database.MediaFile, error) {
	file, err := s.db.GetMediaFile(ctx, fileID)
	if err != nil {
		return nil, nil, ErrFileNotFound
	}

	if file.ScanStatus == "flagged" {
		return nil, nil, ErrFileFlagged
	}

	return nil, file, nil
}

// GetChunks returns chunk metadata for a file.
func (s *Service) GetChunks(ctx context.Context, fileID string) ([]*database.MediaChunk, error) {
	file, err := s.db.GetMediaFile(ctx, fileID)
	if err != nil {
		return nil, ErrFileNotFound
	}
	_ = file
	return s.db.GetChunks(ctx, fileID)
}

// SubmitForScan submits a file for virus/content scanning.
func (s *Service) SubmitForScan(ctx context.Context, fileID string) error {
	_, err := s.db.GetMediaFile(ctx, fileID)
	if err != nil {
		return ErrFileNotFound
	}
	// In production, this would submit to an async scanning pipeline.
	return s.db.UpdateScanStatus(ctx, fileID, "scanning")
}

// MemoryStorage is an in-memory StorageBackend for testing.
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewMemoryStorage creates a new in-memory storage backend.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string][]byte)}
}

func (m *MemoryStorage) Store(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = data
	return nil
}

func (m *MemoryStorage) Retrieve(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.data[key]
	if !ok {
		return nil, ErrFileNotFound
	}
	return d, nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}
