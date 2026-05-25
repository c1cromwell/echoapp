package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// TestTrustTierCommitment_AnchorsHashOnly verifies D4: the handler anchors
// H(tier||nonce) on the Identity Metagraph and never sends the raw tier/nonce
// on-chain (raw trust scores stay off-chain).
func TestTrustTierCommitment_AnchorsHashOnly(t *testing.T) {
	var onChain []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		onChain, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hash":"deadbeef"}`))
	}))
	defer srv.Close()

	rt := &Router{
		IdentityL1: metagraph.NewMetagraphClient(metagraph.MetagraphConfig{IdentityL1URL: srv.URL}),
	}

	body, _ := json.Marshal(map[string]interface{}{"tier": 3, "nonce": "secret-nonce-xyz"})
	req := httptest.NewRequest(http.MethodPost, "/identity/trust-tier/commitment", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zUser"))
	rec := httptest.NewRecorder()
	rt.handleTrustTierCommitment(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	want := metagraph.TrustTierCommitmentHex(3, "secret-nonce-xyz")
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["commitment"] != want {
		t.Fatalf("response commitment = %v, want %s", resp["commitment"], want)
	}

	// The on-chain payload must carry the commitment but never the raw nonce/tier preimage.
	if !bytes.Contains(onChain, []byte(want)) {
		t.Fatalf("on-chain payload missing commitment: %s", onChain)
	}
	if bytes.Contains(onChain, []byte("secret-nonce-xyz")) {
		t.Fatalf("raw nonce leaked on-chain: %s", onChain)
	}
}

func TestTrustTierCommitment_Validation(t *testing.T) {
	rt := &Router{
		IdentityL1: metagraph.NewMetagraphClient(metagraph.MetagraphConfig{IdentityL1URL: "http://unused"}),
	}
	call := func(payload map[string]interface{}) int {
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/identity/trust-tier/commitment", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zUser"))
		rec := httptest.NewRecorder()
		rt.handleTrustTierCommitment(rec, req)
		return rec.Code
	}
	if c := call(map[string]interface{}{"tier": 9, "nonce": "n"}); c != http.StatusBadRequest {
		t.Fatalf("out-of-range tier want 400, got %d", c)
	}
	if c := call(map[string]interface{}{"tier": 3}); c != http.StatusBadRequest {
		t.Fatalf("missing nonce want 400, got %d", c)
	}
}

func TestTrustTierCommitment_NoL1(t *testing.T) {
	rt := &Router{} // IdentityL1 nil
	body, _ := json.Marshal(map[string]interface{}{"tier": 3, "nonce": "n"})
	req := httptest.NewRequest(http.MethodPost, "/identity/trust-tier/commitment", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), ContextKeyUserID, "did:key:zUser"))
	rec := httptest.NewRecorder()
	rt.handleTrustTierCommitment(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 without Identity L1, got %d", rec.Code)
	}
}
