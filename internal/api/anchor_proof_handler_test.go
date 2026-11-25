package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/ledger"
)

func TestHandleAnchorProofBundle(t *testing.T) {
	ctx := context.Background()
	store := ledger.NewMemoryProofStore()
	leaves := []string{strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)}
	tree := ledger.BuildMerkleTree(leaves)
	root := tree.Root
	_ = store.RecordRootSubmitted(ctx, root, len(leaves), "besu", time.Now())
	_ = store.MarkRootFinalized(ctx, root, time.Now())
	_ = store.SaveChainAnchor(ctx, root, ledger.ChainAnchor{BlockNumber: 42, TxHash: "0xtx", Contract: "0xc0", Proof: json.RawMessage(`{"root":"0x00"}`)})
	_ = store.Put(ctx, ledger.MessageAnchorProof{MessageID: "msg-2", Commitment: leaves[2], Siblings: tree.ProofSiblings(2), MerkleLeafIndex: 2, MerkleRoot: root})

	// No chain clients, no anchoring worker: the durable record is enough.
	rt := NewRouter(nil)
	rt.AnchorStore = store
	rt.ChainNetwork = "besu"

	do := func(method, target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		rt.handleAnchorProofBundle(rec, httptest.NewRequest(method, target, nil))
		return rec
	}

	rec := do(http.MethodGet, "/v1/anchors/"+root+"/proof?messageId=msg-2")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var b ledger.ProofBundle
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if b.Network != "besu" || b.MerkleRoot != root || !b.Finalized || b.LeafCount != 3 {
		t.Fatalf("bundle header: %+v", b)
	}
	if b.Chain == nil || b.Chain.BlockNumber != 42 || b.Chain.TxHash != "0xtx" || b.Chain.Contract != "0xc0" || len(b.Chain.Proof) == 0 {
		t.Fatalf("chain evidence: %+v", b.Chain)
	}
	if b.Message == nil || b.Message.MessageID != "msg-2" {
		t.Fatalf("message: %+v", b.Message)
	}
	if err := ledger.VerifyMerklePath(b.Message.Commitment, b.Message.Siblings, b.Message.LeafIndex, b.LeafCount, b.MerkleRoot); err != nil {
		t.Fatalf("served path does not verify: %v", err)
	}

	// Without messageId there is no message section.
	rec = do(http.MethodGet, "/v1/anchors/"+root+"/proof")
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"message"`) {
		t.Fatalf("no-message bundle: %d %s", rec.Code, rec.Body.String())
	}

	for name, tc := range map[string]struct {
		method, target string
		want           int
	}{
		"unknown root":    {http.MethodGet, "/v1/anchors/" + strings.Repeat("0", 64) + "/proof", http.StatusNotFound},
		"unknown message": {http.MethodGet, "/v1/anchors/" + root + "/proof?messageId=nope", http.StatusNotFound},
		"malformed root":  {http.MethodGet, "/v1/anchors/not-a-root/proof", http.StatusBadRequest},
		"uppercase root":  {http.MethodGet, "/v1/anchors/" + strings.Repeat("A", 64) + "/proof", http.StatusBadRequest},
		"nested path":     {http.MethodGet, "/v1/anchors/" + root + "/x/proof", http.StatusBadRequest},
		"wrong method":    {http.MethodPost, "/v1/anchors/" + root + "/proof", http.StatusMethodNotAllowed},
	} {
		if rec := do(tc.method, tc.target); rec.Code != tc.want {
			t.Errorf("%s: status %d, want %d (%s)", name, rec.Code, tc.want, rec.Body.String())
		}
	}

	rt.AnchorStore = nil
	if rec := do(http.MethodGet, "/v1/anchors/"+root+"/proof"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("no store: status %d", rec.Code)
	}
}

// The route is reachable through the real handler chain and requires auth,
// like GET /v1/messages/{id}/merkle-proof.
func TestAnchorProofRouteRequiresAuth(t *testing.T) {
	rt := NewRouter(nil)
	rt.AnchorStore = ledger.NewMemoryProofStore()
	for _, path := range []string{
		"/v1/anchors/" + strings.Repeat("0", 64) + "/proof",
		"/v1/messages/msg-1/merkle-proof",
	} {
		rec := httptest.NewRecorder()
		rt.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without a token: status %d, want 401", path, rec.Code)
		}
	}
}

// The per-message proof endpoint reads the durable store even when the
// anchoring worker (and so the chain) is not configured.
func TestHandleMessageMerkleProof_DurableStoreWithoutAnchoring(t *testing.T) {
	store := ledger.NewMemoryProofStore()
	_ = store.Put(context.Background(), ledger.MessageAnchorProof{MessageID: "msg-7", Commitment: "aa", MerkleRoot: "root"})
	rt := NewRouter(nil)
	rt.AnchorStore = store
	rec := httptest.NewRecorder()
	rt.handleMessageMerkleProof(rec, httptest.NewRequest(http.MethodGet, "/v1/messages/msg-7/merkle-proof", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
