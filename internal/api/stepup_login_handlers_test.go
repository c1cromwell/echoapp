package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withUserID(ctx context.Context, did string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, did)
}

func TestLoginChallenge_HappyPath(t *testing.T) {
	rt := &Router{DIDRegistry: NewMemoryDIDRegistry()}
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login/challenge", nil)
	w := httptest.NewRecorder()

	rt.handleLoginChallenge(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["challenge"] == "" {
		t.Error("expected non-empty challenge")
	}
	if int(resp["timeout"].(float64)) != 60 {
		t.Errorf("unexpected timeout: %v", resp["timeout"])
	}
}

func TestLoginChallenge_MethodNotAllowed(t *testing.T) {
	rt := &Router{DIDRegistry: NewMemoryDIDRegistry()}
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/login/challenge", nil)
	w := httptest.NewRecorder()
	rt.handleLoginChallenge(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestStepUp_HappyPath(t *testing.T) {
	rt, err := newTestRouterWithTokenService()
	if err != nil {
		t.Skipf("token service unavailable: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"action": "revoke_device"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/step-up", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Inject DID into context (normally set by authMiddleware after signature verification)
	req = req.WithContext(withUserID(req.Context(), "did:key:z6MkTestStepUp"))
	w := httptest.NewRecorder()

	rt.handleStepUp(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["elevated_token"] == "" {
		t.Error("expected non-empty elevated_token")
	}
	if resp["action"] != "revoke_device" {
		t.Errorf("unexpected action: %v", resp["action"])
	}
}

func TestStepUp_UnknownAction(t *testing.T) {
	rt, err := newTestRouterWithTokenService()
	if err != nil {
		t.Skipf("token service unavailable: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"action": "something_invalid"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/step-up", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withUserID(req.Context(), "did:key:z6MkTest"))
	w := httptest.NewRecorder()

	rt.handleStepUp(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestStepUp_MissingAction(t *testing.T) {
	rt, err := newTestRouterWithTokenService()
	if err != nil {
		t.Skipf("token service unavailable: %v", err)
	}
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/step-up", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withUserID(req.Context(), "did:key:z6MkTest"))
	w := httptest.NewRecorder()

	rt.handleStepUp(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestStepUp_Unauthenticated(t *testing.T) {
	rt, err := newTestRouterWithTokenService()
	if err != nil {
		t.Skipf("token service unavailable: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"action": "revoke_device"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/step-up", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No DID in context — unauthenticated
	w := httptest.NewRecorder()

	rt.handleStepUp(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// newTestRouterWithTokenService creates a Router with a real TokenService for tests.
func newTestRouterWithTokenService() (*Router, error) {
	rt := NewRouter([]string{"*"})
	return rt, nil
}
