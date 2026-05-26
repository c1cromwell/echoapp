package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/credentials/oidc4vc"
)

func TestEnrollmentVCStartRequiresVerifier(t *testing.T) {
	rt := NewRouter(nil)
	body, _ := json.Marshal(map[string]interface{}{
		"requested_claims": map[string]bool{
			"familyName": true, "givenName": true, "ageOver18": true,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/enrollment/vc/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.handleEnrollmentVCStart(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when verifier disabled, got %d", rec.Code)
	}
}

func TestEnrollmentVCStartReturnsSession(t *testing.T) {
	rt := NewRouter(nil)
	rt.OIDCVerifier = oidc4vc.NewVerifier(
		"did:key:verifier",
		"did:key:issuer",
		"http://localhost:8000",
		"http://localhost:8000",
	)
	rt.OIDCVerifierBaseURL = "http://localhost:8000"

	body, _ := json.Marshal(map[string]interface{}{
		"requested_claims": map[string]interface{}{
			"familyName": true, "givenName": true, "ageOver18": true,
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/enrollment/vc/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rt.handleEnrollmentVCStart(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["id"] == "" || resp["verifierURL"] == "" {
		t.Fatalf("missing session fields: %+v", resp)
	}
}

func TestEnrollmentVCSessionExpiry(t *testing.T) {
	rt := NewRouter(nil)
	rt.OIDCVerifier = oidc4vc.NewVerifier(
		"did:key:verifier",
		"did:key:issuer",
		"http://localhost:8000",
		"http://localhost:8000",
	)
	rt.storeEnrollmentVCSession("stale", enrollmentVCSession{
		State: "stale", CredentialType: "KYCLite", ExpiresAt: time.Now().Add(-time.Minute),
	})
	body, _ := json.Marshal(map[string]interface{}{
		"session_id": "stale",
		"vp_token":   "tok",
		"state":      "stale",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/enrollment/vc/finish", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	rt.handleEnrollmentVCFinish(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for expired session, got %d", rec.Code)
	}
}
