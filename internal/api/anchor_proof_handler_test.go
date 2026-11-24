package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func TestHandleAnchorProofBundle(t *testing.T) {
	ctx := context.Background()
	store := metagraph.NewMemoryProofStore()
	leaves := []string{strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)}
	tree := metagraph.BuildMerkleTree(leaves)
	root := tree.Root
	_ = store.RecordRootSubmitted(ctx, root, len(leaves), "sovereign", time.Now())
	_ = store.MarkRootFinalized(ctx, root, time.Now())
	yes := true
	_ = store.SaveRootSnapshots(ctx, root,
		&metagraph.SnapshotRef{Ordinal: 42, Hash: "mh", Signers: []string{"p1", "p2"}, ArchiveKey: "chain-archive/sovereign/ml0/42-mh.json", SHA256: "s1", ContainsRoot: &yes},
		&metagraph.SnapshotRef{Ordinal: 900, Hash: "gh", Signers: []string{"g1"}, ArchiveKey: "chain-archive/sovereign/gl0/900-gh.json", SHA256: "s2"})
	_ = store.Put(ctx, metagraph.MessageAnchorProof{MessageID: "msg-2", Commitment: leaves[2], Siblings: tree.ProofSiblings(2), MerkleLeafIndex: 2, MerkleRoot: root})

	// No chain clients, no anchoring worker: the durable record is enough.
	rt := NewRouter(nil)
	rt.AnchorStore = store
	rt.ChainNetwork = "sovereign"

	do := func(method, target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		rt.handleAnchorProofBundle(rec, httptest.NewRequest(method, target, nil))
		return rec
	}

	rec := do(http.MethodGet, "/v1/anchors/"+root+"/proof?messageId=msg-2")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var b metagraph.ProofBundle
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if b.Network != "sovereign" || b.MerkleRoot != root || !b.Finalized || b.LeafCount != 3 {
		t.Fatalf("bundle header: %+v", b)
	}
	if b.MetagraphSnapshot == nil || b.MetagraphSnapshot.Ordinal != 42 || b.MetagraphSnapshot.Hash != "mh" ||
		len(b.MetagraphSnapshot.Signers) != 2 || b.MetagraphSnapshot.SHA256 != "s1" ||
		b.MetagraphSnapshot.ArchiveKey != "chain-archive/sovereign/ml0/42-mh.json" {
		t.Fatalf("metagraph snapshot: %+v", b.MetagraphSnapshot)
	}
	if b.GlobalSnapshot == nil || b.GlobalSnapshot.Ordinal != 900 || b.GlobalSnapshot.Signers[0] != "g1" {
		t.Fatalf("global snapshot: %+v", b.GlobalSnapshot)
	}
	if b.Message == nil || b.Message.MessageID != "msg-2" {
		t.Fatalf("message: %+v", b.Message)
	}
	if err := metagraph.VerifyMerklePath(b.Message.Commitment, b.Message.Siblings, b.Message.LeafIndex, b.LeafCount, b.MerkleRoot); err != nil {
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
	rt.AnchorStore = metagraph.NewMemoryProofStore()
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
	store := metagraph.NewMemoryProofStore()
	_ = store.Put(context.Background(), metagraph.MessageAnchorProof{MessageID: "msg-7", Commitment: "aa", MerkleRoot: "root"})
	rt := NewRouter(nil)
	rt.AnchorStore = store
	rec := httptest.NewRecorder()
	rt.handleMessageMerkleProof(rec, httptest.NewRequest(http.MethodGet, "/v1/messages/msg-7/merkle-proof", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
