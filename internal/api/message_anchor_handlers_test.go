package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/ledger"
)

func TestHandleMessageMerkleProof(t *testing.T) {
	mem := ledger.NewMemoryProofStore()
	_ = mem.Put(context.Background(), ledger.MessageAnchorProof{
		MessageID:  "msg-42",
		Commitment: "aa",
		Siblings:   []string{"bb"},
		MerkleRoot: "root",
	})
	rt := NewRouter(nil)
	rt.Anchoring = ledger.NewAnchoringService(ledger.AnchoringConfig{Proofs: mem})

	req := httptest.NewRequest(http.MethodGet, "/v1/messages/msg-42/merkle-proof", nil)
	rec := httptest.NewRecorder()
	rt.handleMessageMerkleProof(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
