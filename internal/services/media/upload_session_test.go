package media

import (
	"context"
	"testing"
)

func TestResumableUpload(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	sess, err := svc.InitUpload(ctx, UploadRequest{
		UploaderDID:   "did:key:alice",
		ContentType:   "image/jpeg",
		EncryptedSize: 300_000,
		TrustTier:     3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess.ChunkCount < 2 {
		t.Fatalf("expected >=2 chunks, got %d", sess.ChunkCount)
	}
	chunk := make([]byte, ChunkSize)
	if err := svc.UploadChunk(ctx, sess.FileID, 0, chunk); err != nil {
		t.Fatal(err)
	}
	if err := svc.UploadChunk(ctx, sess.FileID, 1, chunk[:44_000]); err != nil {
		t.Fatal(err)
	}
	result, err := svc.CompleteUpload(ctx, sess.FileID)
	if err != nil {
		t.Fatal(err)
	}
	if result.ChunkCount != sess.ChunkCount {
		t.Fatalf("chunk count mismatch: %d vs %d", result.ChunkCount, sess.ChunkCount)
	}
}
