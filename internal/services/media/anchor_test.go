package media

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/ledger"
)

var hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// TestUpload_AnchorsContentRootOnDataL1 verifies D3 CID anchoring: an upload to a
// content-addressed backend anchors a SHA-256 Merkle root over the chunk CIDs on
// Data L1, and surfaces that root to the caller.
func TestUpload_AnchorsContentRootOnDataL1(t *testing.T) {
	var update ledger.DataL1MerkleRootUpdate
	chain := ledger.SubmitterFunc(func(_ context.Context, tx interface{}) (string, error) {
		update = tx.(ledger.DataL1MerkleRootUpdate)
		return "0xtx", nil
	})

	svc := NewService(database.NewMemoryDB(), NewMemoryStorage()) // MemoryStorage returns content hashes as CIDs
	svc.DataL1 = chain

	res, err := svc.Upload(context.Background(), UploadRequest{
		UploaderDID:   "did:key:zUploader",
		ContentType:   "image/png",
		EncryptedSize: 11,
		TrustTier:     3,
	}, bytes.NewReader([]byte("hello world")))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if !hex64.MatchString(res.ContentRoot) {
		t.Fatalf("ContentRoot should be a 64-hex Merkle root, got %q", res.ContentRoot)
	}

	if update.Root != res.ContentRoot {
		t.Fatalf("anchored root %q != result root %q", update.Root, res.ContentRoot)
	}
	if update.LeafCount != res.ChunkCount || update.LeafCount == 0 {
		t.Fatalf("leafCount %d should equal chunk count %d (non-zero)", update.LeafCount, res.ChunkCount)
	}
}

// TestUpload_NoAnchorWithoutDataL1 confirms uploads succeed (and skip anchoring)
// when no Data L1 client is configured.
func TestUpload_NoAnchorWithoutDataL1(t *testing.T) {
	svc := NewService(database.NewMemoryDB(), NewMemoryStorage()) // no DataL1
	res, err := svc.Upload(context.Background(), UploadRequest{
		UploaderDID:   "did:key:zUploader",
		ContentType:   "image/png",
		EncryptedSize: 5,
		TrustTier:     3,
	}, bytes.NewReader([]byte("hi")))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	// MemoryStorage yields CIDs, so a content root is still computed and returned,
	// even though nothing was anchored on-chain.
	if !hex64.MatchString(res.ContentRoot) {
		t.Fatalf("expected a content root even without anchoring, got %q", res.ContentRoot)
	}
}
