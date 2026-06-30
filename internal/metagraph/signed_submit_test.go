package metagraph

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// SubmitSignedTransaction must relay the client-signed {value, proofs} payload
// verbatim to the right endpoint and return the metagraph's tx hash. The backend
// is a pure relay here — it never signs.
func TestSubmitSignedTransactionRelaysPayload(t *testing.T) {
	var gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hash":"tokenlock_hash_1"}`))
	}))
	defer srv.Close()

	client := NewMetagraphClient(MetagraphConfig{CurrencyL1URL: srv.URL})
	signed := json.RawMessage(`{"value":{"amount":100000000,"source":"DAGxyz"},"proofs":[{"id":"abcd","signature":"3045"}]}`)

	hash, err := client.SubmitSignedTransaction(context.Background(), srv.URL, "/token-locks", signed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "tokenlock_hash_1" {
		t.Fatalf("hash = %q", hash)
	}
	if gotPath != "/token-locks" {
		t.Fatalf("path = %q, want /token-locks", gotPath)
	}
	// Payload must be forwarded unmodified (with the proofs intact).
	if !strings.Contains(string(gotBody), `"proofs"`) || !strings.Contains(string(gotBody), `"signature":"3045"`) {
		t.Fatalf("forwarded body missing signed proofs: %s", gotBody)
	}
}

// Withdraw routes to the Global L0 delegated-stakes endpoint as a PUT.
func TestSubmitSignedByTypeWithdrawUsesPut(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"hash":"wd_1"}`))
	}))
	defer srv.Close()

	client := NewMetagraphClient(MetagraphConfig{L0URL: srv.URL})
	hash, err := client.SubmitSignedByType(context.Background(), "withdrawDelegatedStake",
		[]byte(`{"value":{"stakeRef":"r1"},"proofs":[{"id":"a","signature":"30"}]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "wd_1" || gotMethod != http.MethodPut || gotPath != "/delegated-stakes" {
		t.Fatalf("withdraw routing wrong: hash=%q method=%q path=%q", hash, gotMethod, gotPath)
	}
}
