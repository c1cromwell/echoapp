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
	// Inject DID into context and the sender header, as authMiddleware leaves
	// them after verifying a passkey signature.
	req = req.WithContext(withUserID(req.Context(), "did:key:z6MkTestStepUp"))
	req.Header.Set(headerSenderDID, "did:key:z6MkTestStepUp")
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
	req.Header.Set(headerSenderDID, "did:key:z6MkTest")
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
	req.Header.Set(headerSenderDID, "did:key:z6MkTest")
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

// A session (Bearer) token alone must not be able to mint an elevated token:
// step-up has to be signed by the device key.
func TestStepUp_BearerOnlyIsRefused(t *testing.T) {
	rt := NewRouter([]string{"*"})
	access := accessTokenFor(t, rt, "did:key:zStepUpBearer")
	for _, action := range []string{"delete_account", "revoke_device", "export_recovery"} {
		rec := secPostJSON(t, rt, "/v1/auth/step-up", map[string]string{"action": action}, access)
		if rec.Code != http.StatusUnauthorized || !bytes.Contains(rec.Body.Bytes(), []byte("PASSKEY_REQUIRED")) {
			t.Fatalf("%s: want 401 PASSKEY_REQUIRED, got %d %s", action, rec.Code, rec.Body.String())
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("elevated_token")) {
			t.Fatalf("%s: no elevated token may be issued to a Bearer-only caller", action)
		}
	}
}

// The flow the iOS client uses to delete an account: a passkey-signed step-up
// request yields an elevated token that the delete route accepts.
func TestStepUp_PasskeySignedRequestGetsElevatedToken(t *testing.T) {
	rt := NewRouter([]string{"*"})
	if rt.stepUp == nil {
		t.Fatal("NewRouter must set the step-up service")
	}
	dev := enrollTestDevice(t, rt)

	body, _ := json.Marshal(map[string]string{"action": "delete_account"})
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, dev.signed(t, http.MethodPost, "/v1/auth/step-up", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("signed step-up: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		ElevatedToken string `json:"elevated_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.ElevatedToken == "" {
		t.Fatalf("no elevated token: %v %s", err, rec.Body.String())
	}
	if !rt.elevatedFor(elevatedRequest(out.ElevatedToken), dev.did, "delete_account") {
		t.Fatal("the issued token should be elevated for delete_account")
	}
	if rt.elevatedFor(elevatedRequest(out.ElevatedToken), dev.did, "revoke_device") {
		t.Fatal("the issued token must not cover another action")
	}
}

func elevatedRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, "/v1/users/account", nil)
	req.Header.Set("X-Elevated-Token", token)
	return req
}

// newTestRouterWithTokenService creates a Router with a real TokenService for tests.
func newTestRouterWithTokenService() (*Router, error) {
	rt := NewRouter([]string{"*"})
	return rt, nil
}
