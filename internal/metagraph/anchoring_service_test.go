package metagraph_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/services/relay"
)

type stubCommitments struct {
	entries []relay.CommitmentEntry
}

func (s *stubCommitments) FlushCommitments() []relay.CommitmentEntry {
	out := s.entries
	s.entries = nil
	return out
}

type stubConfirm struct {
	calls int
}

func (s *stubConfirm) PublishConfirmation(_ string, _ metagraph.AnchorConfirmation) bool {
	s.calls++
	return true
}

func TestAnchoringService_SubmitsBatchAndStoresProof(t *testing.T) {
	var submitted metagraph.DataL1MerkleRootUpdate
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&submitted)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"hash":"snap-abc"}`))
	}))
	defer srv.Close()

	client := metagraph.NewMetagraphClient(metagraph.MetagraphConfig{
		DataL1URL: srv.URL,
	})
	src := &stubCommitments{entries: []relay.CommitmentEntry{
		{MessageID: "m1", SenderDID: "did:key:alice", Hash: []byte("hash-one")},
		{MessageID: "m2", SenderDID: "did:key:bob", Hash: []byte("hash-two")},
	}}
	confirm := &stubConfirm{}
	svc := metagraph.NewAnchoringService(metagraph.AnchoringConfig{
		DataL1:    client,
		Source:    src,
		Confirm:   confirm,
		Proofs:    metagraph.NewMemoryProofStore(),
		Network:   "besu",
		LastFlush: time.Now().Add(-10 * time.Minute),
	})

	svc.Tick(context.Background())

	if submitted.LeafCount != 2 {
		t.Fatalf("leaf count = %d, want 2", submitted.LeafCount)
	}
	if submitted.Root == "" {
		t.Fatal("expected merkle root submission")
	}
	proof, ok, err := svc.ProofStore().Get(context.Background(), "m1")
	if err != nil || !ok {
		t.Fatalf("proof lookup: ok=%v err=%v", ok, err)
	}
	if proof.SnapshotHash != "snap-abc" {
		t.Fatalf("snapshot hash = %q", proof.SnapshotHash)
	}
	if len(proof.Siblings) == 0 {
		t.Fatal("expected merkle siblings")
	}
	if confirm.calls != 2 {
		t.Fatalf("confirmations = %d, want 2", confirm.calls)
	}
	// The root is recorded as final (the submitter returns only once the chain has it), and the
	// stored path verifies.
	roots := svc.ProofStore().(metagraph.RootStore)
	rec, ok, err := roots.GetRoot(context.Background(), submitted.Root)
	if err != nil || !ok || rec.LeafCount != 2 || rec.Network != "besu" || !rec.Finalized {
		t.Fatalf("anchored root record: %+v ok=%v err=%v", rec, ok, err)
	}
	if err := metagraph.VerifyMerklePath(proof.Commitment, proof.Siblings, proof.MerkleLeafIndex, rec.LeafCount, rec.MerkleRoot); err != nil {
		t.Fatalf("stored proof does not verify: %v", err)
	}
}
