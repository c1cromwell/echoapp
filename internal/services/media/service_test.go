package media

import (
	"bytes"
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func setupTest() (*Service, database.DB) {
	db := database.NewMemoryDB()
	storage := NewMemoryStorage()
	svc := NewService(db, storage)
	return svc, db
}

func TestUpload(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	data := bytes.Repeat([]byte("x"), 1024)
	req := UploadRequest{
		UploaderDID:   "did:alice",
		ContentType:   "image/png",
		EncryptedSize: int64(len(data)),
		TrustTier:     3,
	}

	result, err := svc.Upload(ctx, req, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if result.FileID == "" {
		t.Error("expected non-empty file ID")
	}
	if result.Size != 1024 {
		t.Errorf("expected size 1024, got %d", result.Size)
	}
}

func TestUploadTier1Rejected(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	req := UploadRequest{
		UploaderDID:   "did:alice",
		ContentType:   "image/png",
		EncryptedSize: 1024,
		TrustTier:     1,
	}

	_, err := svc.Upload(ctx, req, bytes.NewReader([]byte("x")))
	if err != ErrTierRestricted {
		t.Errorf("expected ErrTierRestricted, got %v", err)
	}
}

func TestUploadFileTooLarge(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	req := UploadRequest{
		UploaderDID:   "did:alice",
		ContentType:   "image/png",
		EncryptedSize: 200 * 1024 * 1024, // 200MB > 100MB tier 3 limit
		TrustTier:     3,
	}

	_, err := svc.Upload(ctx, req, bytes.NewReader([]byte("x")))
	if err != ErrFileTooLarge {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestUploadInvalidContentType(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	req := UploadRequest{
		UploaderDID:   "did:alice",
		ContentType:   "application/javascript",
		EncryptedSize: 100,
		TrustTier:     3,
	}

	_, err := svc.Upload(ctx, req, bytes.NewReader([]byte("x")))
	if err != ErrInvalidContent {
		t.Errorf("expected ErrInvalidContent, got %v", err)
	}
}

func TestDownload(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.StoreMediaFile(ctx, &database.MediaFile{
		FileID: "file-1", UploaderDID: "did:alice", ContentType: "image/png",
		EncryptedSize: 1024, ScanStatus: "clean",
	})

	_, meta, err := svc.Download(ctx, "file-1", "did:bob")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if meta.ContentType != "image/png" {
		t.Errorf("expected image/png, got %s", meta.ContentType)
	}
}

func TestDownloadNotFound(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	_, _, err := svc.Download(ctx, "nonexistent", "did:bob")
	if err != ErrFileNotFound {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestDownloadFlagged(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.StoreMediaFile(ctx, &database.MediaFile{
		FileID: "file-1", UploaderDID: "did:alice", ScanStatus: "flagged",
	})

	_, _, err := svc.Download(ctx, "file-1", "did:bob")
	if err != ErrFileFlagged {
		t.Errorf("expected ErrFileFlagged, got %v", err)
	}
}

func TestSubmitForScan(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.StoreMediaFile(ctx, &database.MediaFile{
		FileID: "file-1", UploaderDID: "did:alice",
	})

	if err := svc.SubmitForScan(ctx, "file-1"); err != nil {
		t.Fatalf("SubmitForScan: %v", err)
	}

	file, _ := db.GetMediaFile(ctx, "file-1")
	if file.ScanStatus != "scanning" {
		t.Errorf("expected scanning, got %s", file.ScanStatus)
	}
}

func TestSubmitForScanNotFound(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	if err := svc.SubmitForScan(ctx, "nonexistent"); err != ErrFileNotFound {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestGetChunks(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.StoreMediaFile(ctx, &database.MediaFile{FileID: "file-1", UploaderDID: "did:alice"})
	db.StoreChunk(ctx, &database.MediaChunk{ChunkID: "c1", FileID: "file-1", Index: 0, Size: 256})
	db.StoreChunk(ctx, &database.MediaChunk{ChunkID: "c2", FileID: "file-1", Index: 1, Size: 128})

	chunks, err := svc.GetChunks(ctx, "file-1")
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunks))
	}
}

func TestMaxSizeForTier(t *testing.T) {
	if MaxSizeForTier[1] != 0 {
		t.Error("tier 1 should have 0 max size")
	}
	if MaxSizeForTier[2] != 10*1024*1024 {
		t.Error("tier 2 should be 10MB")
	}
	if MaxSizeForTier[5] != 2*1024*1024*1024 {
		t.Error("tier 5 should be 2GB")
	}
}

func TestAllowedContentTypes(t *testing.T) {
	if !AllowedContentTypes["image/jpeg"] {
		t.Error("image/jpeg should be allowed")
	}
	if AllowedContentTypes["application/javascript"] {
		t.Error("application/javascript should not be allowed")
	}
}

func TestMemoryStorage(t *testing.T) {
	ms := NewMemoryStorage()
	ctx := context.Background()

	ms.Store(ctx, "key1", []byte("data"))

	got, err := ms.Retrieve(ctx, "key1")
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("expected 'data', got '%s'", string(got))
	}

	ms.Delete(ctx, "key1")
	_, err = ms.Retrieve(ctx, "key1")
	if err == nil {
		t.Error("expected error after delete")
	}
}
