package media

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

var hex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// TestUpload_AnchorsContentRootOnDataL1 verifies D3 CID anchoring: an upload to a
// content-addressed backend anchors a SHA-256 Merkle root over the chunk CIDs on
// Data L1, and surfaces that root to the caller.
func TestUpload_AnchorsContentRootOnDataL1(t *testing.T) {
	var anchored []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/transactions" && r.Method == http.MethodPost {
			anchored, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"hash":"x"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	svc := NewService(database.NewMemoryDB(), NewMemoryStorage()) // MemoryStorage returns content hashes as CIDs
	svc.DataL1 = metagraph.NewMetagraphClient(metagraph.MetagraphConfig{DataL1URL: srv.URL})

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

	var update metagraph.DataL1MerkleRootUpdate
	if err := json.Unmarshal(anchored, &update); err != nil {
		t.Fatalf("anchored payload not a DataL1MerkleRootUpdate: %s", anchored)
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
