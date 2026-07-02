// Package media implements the encrypted media upload/download service.
package media

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
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
	// ContentRoot is the SHA-256 Merkle root over the chunk CIDs, anchored on
	// Data L1 when content-addressed storage + a DataL1 client are configured.
	ContentRoot string `json:"contentRoot,omitempty"`
}

// StorageBackend abstracts the underlying storage (Storj/S3/IPFS).
type StorageBackend interface {
	// Store persists data under key and returns its content identifier (CID) when
	// the backend is content-addressed (IPFS); content-location backends (S3)
	// return an empty cid. The cid enables on-chain integrity anchoring (D3).
	Store(ctx context.Context, key string, data []byte) (cid string, err error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// Service provides media upload/download operations.
type Service struct {
	db       database.DB
	storage  StorageBackend
	sessions *sessionStore
	// DataL1 optionally anchors a content Merkle root (over chunk CIDs) on the
	// Data L1 metagraph so media integrity/provenance is publicly verifiable (D3).
	// Nil disables anchoring; only content-addressed backends (IPFS) yield CIDs.
	DataL1 *metagraph.MetagraphClient
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
	var cidLeaves []string // sha256(chunkCID) per chunk, for the content Merkle root

	for {
		n, err := body.Read(buf)
		if n > 0 {
			chunkData := make([]byte, n)
			copy(chunkData, buf[:n])

			chunkID := fmt.Sprintf("%s-chunk-%d", fileID, chunkIndex)
			checksum := fmt.Sprintf("%x", sha256.Sum256(chunkData))

			// Store chunk data
			if s.storage != nil {
				cid, err := s.storage.Store(ctx, chunkID, chunkData)
				if err != nil {
					return nil, err
				}
				if cid != "" {
					cidLeaves = append(cidLeaves, fmt.Sprintf("%x", sha256.Sum256([]byte(cid))))
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

	// D3: anchor a content Merkle root over the chunk CIDs on Data L1 so anyone
	// can verify the media's integrity/provenance against the chain. Best-effort:
	// only when storage is content-addressed (CIDs present) and a client is wired.
	contentRoot := s.anchorContentRoot(ctx, fileID, cidLeaves)

	return &UploadResult{
		FileID:      fileID,
		ChunkCount:  chunkIndex,
		ContentType: req.ContentType,
		Size:        totalSize,
		Timestamp:   time.Now(),
		ContentRoot: contentRoot,
	}, nil
}

// anchorContentRoot computes the SHA-256 Merkle root over the chunk CID leaves
// and anchors it on Data L1. Returns the root (even if anchoring is disabled or
// fails) so it can be surfaced to clients; returns "" when there are no CIDs.
func (s *Service) anchorContentRoot(ctx context.Context, fileID string, cidLeaves []string) string {
	if len(cidLeaves) == 0 {
		return ""
	}
	root := metagraph.ComputeMerkleRoot(cidLeaves)
	if s.DataL1 == nil {
		return root
	}
	if _, err := s.DataL1.SubmitDataL1(ctx, metagraph.DataL1MerkleRootUpdate{
		Root:      root,
		LeafCount: len(cidLeaves),
	}); err != nil {
		log.Printf("media: content-root anchor failed for file %s: %v", fileID, err)
	}
	return root
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

// RetrieveChunk returns encrypted chunk bytes for download (M5).
func (s *Service) RetrieveChunk(ctx context.Context, fileID string, index int) ([]byte, error) {
	if index < 0 {
		return nil, ErrFileNotFound
	}
	if _, err := s.db.GetMediaFile(ctx, fileID); err != nil {
		return nil, ErrFileNotFound
	}
	chunks, err := s.db.GetChunks(ctx, fileID)
	if err != nil {
		return nil, err
	}
	var target *database.MediaChunk
	for _, c := range chunks {
		if c.Index == index {
			target = c
			break
		}
	}
	if target == nil {
		return nil, ErrFileNotFound
	}
	if s.storage == nil {
		return nil, ErrFileNotFound
	}
	return s.storage.Retrieve(ctx, target.ChunkID)
}

// StorageArchiver returns the archiving backend when Filecoin is enabled.
func (s *Service) StorageArchiver() (*ArchivingBackend, bool) {
	ab, ok := s.storage.(*ArchivingBackend)
	return ab, ok
}

// RenewFilecoinDeal renews archival for a media file's first chunk.
func (s *Service) RenewFilecoinDeal(ctx context.Context, fileID string, durationDays int) (*FilecoinDeal, error) {
	ab, ok := s.StorageArchiver()
	if !ok {
		return nil, fmt.Errorf("filecoin archiver not configured")
	}
	chunks, err := s.db.GetChunks(ctx, fileID)
	if err != nil || len(chunks) == 0 {
		return nil, ErrFileNotFound
	}
	return ab.RenewDeal(ctx, chunks[0].ChunkID, durationDays)
}

// FilecoinDealForFile returns archival deal metadata when Filecoin backend is enabled (WO-185).
func (s *Service) FilecoinDealForFile(ctx context.Context, fileID string) *FilecoinDeal {
	ab, ok := s.storage.(*ArchivingBackend)
	if !ok {
		return nil
	}
	chunks, err := s.db.GetChunks(ctx, fileID)
	if err != nil || len(chunks) == 0 {
		return nil
	}
	return ab.DealForKey(chunks[0].ChunkID)
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

func (m *MemoryStorage) Store(ctx context.Context, key string, data []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = data
	// Return a content hash as a stand-in CID so content-addressed code paths
	// (e.g. Data L1 anchoring) are exercisable against the in-memory backend.
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
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
