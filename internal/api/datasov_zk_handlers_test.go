package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/zk"
)

func TestPQModeStatus_Disabled(t *testing.T) {
	t.Setenv("ECHO_PQ_ENABLED", "false")
	h := &V3Handlers{}
	mux := http.NewServeMux()
	h.WireDataSovZK(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/pq/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"available":false`) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestPQModeStatus_Enabled(t *testing.T) {
	t.Setenv("ECHO_PQ_ENABLED", "true")
	h := &V3Handlers{}
	mux := http.NewServeMux()
	h.WireDataSovZK(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/pq/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"available":true`) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestZKVerify_MidnightEnvelope(t *testing.T) {
	h := &V3Handlers{ZKVerifier: zk.NewVerifier()}
	mux := http.NewServeMux()
	h.WireDataSovZK(mux)
	proof := zk.BuildMidnightEnvelope("did:key:alice", "tier3", "n1", "p1")
	rec := postJSON(mux, "/v3/zk/verify", "did:key:alice", map[string]any{
		"subject_did": "did:key:alice",
		"claim_type":  "tier3",
		"proof":       proof,
		"nonce":       "n1",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"verified":true`) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}
