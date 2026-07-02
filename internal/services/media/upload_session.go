package media

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// UploadSession tracks resumable chunked uploads (WO-21 / WO-34).
type UploadSession struct {
	FileID        string
	UploaderDID   string
	ContentType   string
	TrustTier     int
	EncryptedSize int64
	ChunkCount    int
	Received      map[int]bool
	CreatedAt     time.Time
}

type sessionStore struct {
	mu        sync.Mutex
	sessions  map[string]*UploadSession
	chunkData map[string]map[int][]byte // fileID -> index -> bytes (memory MVP)
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions:  make(map[string]*UploadSession),
		chunkData: make(map[string]map[int][]byte),
	}
}

// InitUpload starts a resumable upload session.
func (s *Service) InitUpload(ctx context.Context, req UploadRequest) (*UploadSession, error) {
	maxSize, ok := MaxSizeForTier[req.TrustTier]
	if !ok || maxSize == 0 {
		return nil, ErrTierRestricted
	}
	if req.EncryptedSize > maxSize {
		return nil, ErrFileTooLarge
	}
	if !AllowedContentTypes[req.ContentType] {
		return nil, ErrInvalidContent
	}
	if s.sessions == nil {
		s.sessions = newSessionStore()
	}
	fileID := uuid.New().String()
	chunkCount := int((req.EncryptedSize + ChunkSize - 1) / ChunkSize)
	if chunkCount == 0 && req.EncryptedSize > 0 {
		chunkCount = 1
	}
	sess := &UploadSession{
		FileID:        fileID,
		UploaderDID:   req.UploaderDID,
		ContentType:   req.ContentType,
		TrustTier:     req.TrustTier,
		EncryptedSize: req.EncryptedSize,
		ChunkCount:    chunkCount,
		Received:      make(map[int]bool),
		CreatedAt:     time.Now(),
	}
	s.sessions.mu.Lock()
	s.sessions.sessions[fileID] = sess
	s.sessions.chunkData[fileID] = make(map[int][]byte)
	s.sessions.mu.Unlock()
	return sess, nil
}

// UploadChunk stores one chunk for a resumable session.
func (s *Service) UploadChunk(ctx context.Context, fileID string, index int, data []byte) error {
	if s.sessions == nil {
		return ErrFileNotFound
	}
	s.sessions.mu.Lock()
	sess, ok := s.sessions.sessions[fileID]
	if !ok {
		s.sessions.mu.Unlock()
		return ErrFileNotFound
	}
	if index < 0 || index >= sess.ChunkCount {
		s.sessions.mu.Unlock()
		return fmt.Errorf("chunk index out of range")
	}
	s.sessions.chunkData[fileID][index] = append([]byte(nil), data...)
	sess.Received[index] = true
	s.sessions.mu.Unlock()
	return nil
}

// CompleteUpload finalizes a resumable session into media metadata + storage.
func (s *Service) CompleteUpload(ctx context.Context, fileID string) (*UploadResult, error) {
	if s.sessions == nil {
		return nil, ErrFileNotFound
	}
	s.sessions.mu.Lock()
	sess, ok := s.sessions.sessions[fileID]
	chunks := s.sessions.chunkData[fileID]
	if !ok {
		s.sessions.mu.Unlock()
		return nil, ErrFileNotFound
	}
	for i := 0; i < sess.ChunkCount; i++ {
		if !sess.Received[i] {
			s.sessions.mu.Unlock()
			return nil, fmt.Errorf("missing chunk %d", i)
		}
	}
	delete(s.sessions.sessions, fileID)
	delete(s.sessions.chunkData, fileID)
	s.sessions.mu.Unlock()

	var totalSize int64
	var cidLeaves []string
	for i := 0; i < sess.ChunkCount; i++ {
		chunkData := chunks[i]
		n := len(chunkData)
		chunkID := fmt.Sprintf("%s-chunk-%d", fileID, i)
		checksum := fmt.Sprintf("%x", sha256.Sum256(chunkData))
		if s.storage != nil {
			cid, err := s.storage.Store(ctx, chunkID, chunkData)
			if err != nil {
				return nil, err
			}
			if cid != "" {
				cidLeaves = append(cidLeaves, fmt.Sprintf("%x", sha256.Sum256([]byte(cid))))
			}
		}
		chunk := &database.MediaChunk{
			ChunkID:  chunkID,
			FileID:   fileID,
			Index:    i,
			Size:     int64(n),
			Checksum: checksum,
		}
		if err := s.db.StoreChunk(ctx, chunk); err != nil {
			return nil, err
		}
		totalSize += int64(n)
	}

	file := &database.MediaFile{
		FileID:        fileID,
		UploaderDID:   sess.UploaderDID,
		ContentType:   sess.ContentType,
		EncryptedSize: totalSize,
		ChunkCount:    sess.ChunkCount,
		ScanStatus:    "pending",
	}
	if err := s.db.StoreMediaFile(ctx, file); err != nil {
		return nil, err
	}
	contentRoot := s.anchorContentRoot(ctx, fileID, cidLeaves)
	return &UploadResult{
		FileID:      fileID,
		ChunkCount:  sess.ChunkCount,
		ContentType: sess.ContentType,
		Size:        totalSize,
		Timestamp:   time.Now(),
		ContentRoot: contentRoot,
	}, nil
}

// Manifest returns upload progress for resume (WO-34).
func (s *Service) Manifest(ctx context.Context, fileID string) (*UploadSession, error) {
	if s.sessions == nil {
		return nil, ErrFileNotFound
	}
	s.sessions.mu.Lock()
	defer s.sessions.mu.Unlock()
	sess, ok := s.sessions.sessions[fileID]
	if !ok {
		return nil, ErrFileNotFound
	}
	return sess, nil
}
