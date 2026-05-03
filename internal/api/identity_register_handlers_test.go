package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// newTestKey generates a fresh P-256 key pair plus its canonical did:key
// and uncompressed-hex public key for use as test fixtures.
func newTestKey(t *testing.T) (did, pubHex string, priv *ecdsa.PrivateKey) {
	t.Helper()
	p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	d, err := didkey.Derive(&p.PublicKey)
	if err != nil {
		t.Fatalf("derive did:key: %v", err)
	}
	uncompressed := elliptic.Marshal(elliptic.P256(), p.X, p.Y)
	return d, hex.EncodeToString(uncompressed), p
}

// newTestRouter builds a minimal Router suitable for handler tests. It uses
// the in-memory DID registry and accepts every bearer token (auth is not the
// focus of these tests; the /identity/register route is public anyway).
func newTestRouter() *Router {
	rt := &Router{
		AllowedOrigins:  []string{"*"},
		DIDRegistry:     NewMemoryDIDRegistry(),
		TokenValidator:  func(string) bool { return true },
		UserIDExtractor: func(string) string { return "test-user" },
	}
	return rt
}

func doRegister(t *testing.T, rt *Router, body any) *httptest.ResponseRecorder {
	t.Helper()
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/identity/register", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	return w
}

func TestIdentityRegister_HappyPath(t *testing.T) {
	rt := newTestRouter()
	did, pubHex, _ := newTestKey(t)

	w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", w.Code, w.Body.String())
	}

	var resp IdentityRegisterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.DID != did {
		t.Fatalf("response did = %q, want %q", resp.DID, did)
	}
	if resp.PublicKeyHex != pubHex {
		t.Fatalf("response public_key_hex mismatch")
	}
	if resp.Existing {
		t.Fatalf("expected existing=false on first registration")
	}
	if resp.RegisteredAt == "" {
		t.Fatalf("registered_at must be populated")
	}
}

func TestIdentityRegister_Idempotent(t *testing.T) {
	rt := newTestRouter()
	did, pubHex, _ := newTestKey(t)

	if w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex}); w.Code != http.StatusCreated {
		t.Fatalf("first call status = %d, body = %s", w.Code, w.Body.String())
	}

	w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex})
	if w.Code != http.StatusOK {
		t.Fatalf("second call status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	var resp IdentityRegisterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !resp.Existing {
		t.Fatalf("expected existing=true on idempotent re-registration")
	}
}

func TestIdentityRegister_RejectsKeyMismatch(t *testing.T) {
	rt := newTestRouter()
	did1, _, _ := newTestKey(t)
	_, pubHex2, _ := newTestKey(t)

	w := doRegister(t, rt, IdentityRegisterRequest{DID: did1, PublicKeyHex: pubHex2})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "DID_KEY_MISMATCH") {
		t.Fatalf("expected DID_KEY_MISMATCH error, got %s", w.Body.String())
	}
}

func TestIdentityRegister_RejectsConflict(t *testing.T) {
	rt := newTestRouter()
	did, pubHex, _ := newTestKey(t)

	if w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex}); w.Code != http.StatusCreated {
		t.Fatalf("first call status = %d, body = %s", w.Code, w.Body.String())
	}

	// Inject a conflict directly in the registry: same DID, different key hex
	// (going through the handler again with a mismatched key would be caught
	// by the canonical-derivation check first, so we exercise the conflict
	// path at the storage layer here).
	if _, _, err := rt.DIDRegistry.Register(context.Background(), did, "deadbeef"); err == nil {
		t.Fatalf("expected ErrDIDConflict from registry, got nil")
	}
}

func TestIdentityRegister_RejectsMissingFields(t *testing.T) {
	rt := newTestRouter()
	did, pubHex, _ := newTestKey(t)

	cases := []struct {
		name     string
		body     IdentityRegisterRequest
		wantCode string
	}{
		{"missing did", IdentityRegisterRequest{PublicKeyHex: pubHex}, "MISSING_DID"},
		{"missing public_key_hex", IdentityRegisterRequest{DID: did}, "MISSING_PUBLIC_KEY"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRegister(t, rt, tc.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tc.wantCode) {
				t.Fatalf("expected %s, got %s", tc.wantCode, w.Body.String())
			}
		})
	}
}

func TestIdentityRegister_RejectsBadJSON(t *testing.T) {
	rt := newTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/identity/register", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "MALFORMED_BODY") {
		t.Fatalf("expected MALFORMED_BODY, got %s", w.Body.String())
	}
}

func TestIdentityRegister_RejectsWrongMethod(t *testing.T) {
	rt := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/identity/register", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", w.Code)
	}
}

func TestIdentityRegister_PublicNoAuthRequired(t *testing.T) {
	// Build a router that *would* reject any unauthenticated call by default,
	// to prove /identity/register really is in the public-paths bypass.
	rt := newTestRouter()
	rt.TokenValidator = func(string) bool { return false }

	did, pubHex, _ := newTestKey(t)
	w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex})
	if w.Code != http.StatusCreated {
		t.Fatalf("public path should bypass auth; got %d body %s", w.Code, w.Body.String())
	}
}

func TestIdentityResolve_HappyPath(t *testing.T) {
	rt := newTestRouter()
	did, pubHex, _ := newTestKey(t)

	if w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex}); w.Code != http.StatusCreated {
		t.Fatalf("setup register failed: %d %s", w.Code, w.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/identity/"+did, nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	var resp IdentityResolveResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp.DID != did {
		t.Fatalf("response mismatch: %+v", resp)
	}
	if len(resp.Devices) != 1 || resp.Devices[0].PublicKeyHex != pubHex {
		t.Fatalf("devices mismatch: %+v", resp)
	}
}

func TestIdentityResolve_NotFound(t *testing.T) {
	rt := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/identity/did:key:zNotFound", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "DID_NOT_REGISTERED") {
		t.Fatalf("expected DID_NOT_REGISTERED, got %s", w.Body.String())
	}
}

func TestIdentityAddDevice_HappyPath(t *testing.T) {
	rt := newTestRouter()
	subjectDID, pubHex1, priv1 := newTestKey(t)
	_, pubHex2, _ := newTestKey(t)

	if w := doRegister(t, rt, IdentityRegisterRequest{DID: subjectDID, PublicKeyHex: pubHex1}); w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}

	body := IdentityAddDeviceRequest{
		SubjectDID:      subjectDID,
		NewPublicKeyHex: pubHex2,
		DeviceLabel:     "ipad",
		SigningDID:      subjectDID,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(raw)
	sig, err := ecdsa.SignASN1(rand.Reader, priv1, h[:])
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/identity/devices", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(identitySignatureHeader, hex.EncodeToString(sig))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("add device: %d %s", w.Code, w.Body.String())
	}
	var resp IdentityResolveResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Devices) != 2 {
		t.Fatalf("want 2 devices, got %v", resp.Devices)
	}
}

func TestIdentityResolve_PrefixedPath(t *testing.T) {
	rt := newTestRouter()
	did, pubHex, _ := newTestKey(t)
	if w := doRegister(t, rt, IdentityRegisterRequest{DID: did, PublicKeyHex: pubHex}); w.Code != http.StatusCreated {
		t.Fatalf("setup: %d", w.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/identity/resolve/"+did, nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}
