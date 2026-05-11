package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// TestHandleIdentityDeviceToken_RequiresAuth verifies that unauthenticated
// requests to POST /identity/devices/token return 401.
func TestHandleIdentityDeviceToken_RequiresAuth(t *testing.T) {
	rt := NewRouter([]string{"*"})

	req := httptest.NewRequest(http.MethodPost, "/identity/devices/token", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	// No auth header → auth middleware returns 401 before the handler runs.
	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// TestHandleIdentityDeviceToken_NoRedis verifies 503 when Redis is not configured.
func TestHandleIdentityDeviceToken_NoRedis(t *testing.T) {
	rt := NewRouter([]string{"*"})
	// rt.Redis is nil by default in NewRouter.

	ts := rt.TokenService()
	token, _, err := ts.IssueAccessToken("did:key:zTest", "device", 1, "full")
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/identity/devices/token", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("want 503 when Redis not configured, got %d", w.Code)
	}
}

// TestHandleIdentityDeviceToken_MethodNotAllowed verifies only POST is accepted.
func TestHandleIdentityDeviceToken_MethodNotAllowed(t *testing.T) {
	rt := NewRouter([]string{"*"})
	ts := rt.TokenService()
	token, _, _ := ts.IssueAccessToken("did:key:zTest", "device", 1, "full")

	req := httptest.NewRequest(http.MethodGet, "/identity/devices/token", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", w.Code)
	}
}

// TestHandleIdentityListDevices_Returns200 verifies GET /identity/devices/{did}
// returns a JSON array for a registered DID.
func TestHandleIdentityListDevices_Returns200(t *testing.T) {
	rt := NewRouter([]string{"*"})

	// Register a DID first.
	privHex, _ := hex.DecodeString("eaff1084e1322774ce79ad393aa9d925d73473a9de5fe6a50d969672ac66be4f")
	pubHex, _ := hex.DecodeString("047bfc587ef5617b74f66c8c26765adfc6ac311be92ec5546f146b28a026e96edd56e82054fa9de5114d4e59938f29dbe2f63bffa40a321f0f472f28ea30f8d1f4")
	_ = privHex
	did, err := didkey.DeriveFromPublicKeyHex(hex.EncodeToString(pubHex))
	if err != nil {
		t.Fatalf("DeriveFromPublicKeyHex: %v", err)
	}

	// Register via the registry directly (no HTTP round-trip needed).
	if _, _, err := rt.DIDRegistry.Register(nil, did, hex.EncodeToString(pubHex)); err != nil {
		t.Fatalf("Register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/identity/devices/"+did, nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var devices []map[string]string
	if err := json.NewDecoder(w.Body).Decode(&devices); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(devices) == 0 {
		t.Error("expected at least one device in the list")
	}
}

// TestHandleIdentityListDevices_UnknownDID returns 404.
func TestHandleIdentityListDevices_UnknownDID(t *testing.T) {
	rt := NewRouter([]string{"*"})
	req := httptest.NewRequest(http.MethodGet, "/identity/devices/did:key:zNonExistent", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

// TestTryTokenBasedDeviceAdd_NoToken verifies that a body without a "token"
// field falls through to the signature path (returns false).
func TestTryTokenBasedDeviceAdd_FallsThrough(t *testing.T) {
	rt := NewRouter([]string{"*"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/identity/devices", nil)

	// Body without token field — should not be handled by the token path.
	handled := rt.tryTokenBasedDeviceAdd(w, r, []byte(`{"subject_did":"did:key:z1"}`))
	if handled {
		t.Error("non-token body should not be handled by tryTokenBasedDeviceAdd")
	}
}

// TestV3HandleAuthVerify_ReturnsDID verifies that an authenticated request
// to POST /v3/auth/verify returns the caller's DID.
func TestV3HandleAuthVerify_ReturnsDID(t *testing.T) {
	rt := NewRouter([]string{"*"})
	rt.V3 = &V3Handlers{}

	ts := rt.TokenService()
	const testDID = "did:key:zVerifyTest"
	token, _, err := ts.IssueAccessToken(testDID, "device", 1, "full")
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v3/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["did"] != testDID {
		t.Errorf("did: want %q, got %v", testDID, resp["did"])
	}
	if resp["verified"] != true {
		t.Errorf("verified: want true, got %v", resp["verified"])
	}
}
