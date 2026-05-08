package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPhase1TrustTierCommitment_RequiresDevelopmentEnv(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	rt := newTestRouter()
	body, _ := json.Marshal(map[string]any{
		"subject_did": "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		"tier":        2,
		"nonce":       "abc",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/phase1/trust-tier-commitment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", w.Code, w.Body.String())
	}
}

func TestPhase1TrustTierCommitment_RequiresIdentityL1(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	t.Cleanup(func() { _ = os.Unsetenv("ENVIRONMENT") })
	rt := newTestRouter()
	body, _ := json.Marshal(map[string]any{
		"subject_did": "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		"tier":        2,
		"nonce":       "abc",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/phase1/trust-tier-commitment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body = %s", w.Code, w.Body.String())
	}
}
