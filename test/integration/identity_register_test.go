package integration

// Black-box integration tests for POST /identity/register and
// GET /identity/{did} (WO-278). These exercise the real api.Router
// over a real HTTP listener via testutil.StartTestServer — the same
// path the iOS client and scripts/validate-phase1.sh hit.
//
// The unit-level tests at internal/api/identity_register_handlers_test.go
// cover the in-memory DIDRegistry contract directly (idempotency,
// conflict, lookup-not-found). These tests cover the wire format,
// HTTP status codes, route plumbing, and OpenAPI contract.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/testutil"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// p256Key returns a fresh P-256 keypair plus its SEC1 uncompressed hex
// encoding and canonical did:key. Centralised so the assertions below
// match the exact (DID, public_key_hex) pair the server will re-derive.
func p256Key(t *testing.T) (did, pubHex string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate p256 key: %v", err)
	}
	pubHex = hex.EncodeToString(elliptic.Marshal(elliptic.P256(), priv.X, priv.Y))
	did, err = didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatalf("derive did:key: %v", err)
	}
	return did, pubHex
}

// TestIdentityRegister_HappyPath covers the WO-278 / WO-230 Step 2
// canonical flow: derive a did:key locally, POST it together with the
// public key, expect 201 + the binding echoed back.
func TestIdentityRegister_HappyPath(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	did, pubHex := p256Key(t)

	resp := ts.Post("/identity/register", "", map[string]string{
		"did":            did,
		"public_key_hex": pubHex,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)

	if body["did"] != did {
		t.Errorf("response.did = %v, want %s", body["did"], did)
	}
	if body["public_key_hex"] != pubHex {
		t.Errorf("response.public_key_hex = %v, want %s", body["public_key_hex"], pubHex)
	}
	if body["existing"] != false {
		t.Errorf("response.existing = %v, want false on first registration", body["existing"])
	}
	if body["registered_at"] == nil || body["registered_at"] == "" {
		t.Error("response.registered_at must be set")
	}
}

// TestIdentityRegister_Idempotent re-posts the same (did, public_key_hex)
// and expects 200 OK with existing=true rather than 409.
func TestIdentityRegister_Idempotent(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	did, pubHex := p256Key(t)
	body := map[string]string{"did": did, "public_key_hex": pubHex}

	first := ts.Post("/identity/register", "", body)
	first.Body.Close()
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d", first.StatusCode)
	}

	second := ts.Post("/identity/register", "", body)
	defer second.Body.Close()
	if second.StatusCode != http.StatusOK {
		t.Fatalf("idempotent re-register: expected 200, got %d", second.StatusCode)
	}

	var resp map[string]interface{}
	ts.DecodeJSON(second, &resp)
	if resp["existing"] != true {
		t.Errorf("response.existing = %v, want true on idempotent re-register", resp["existing"])
	}
}

// TestIdentityRegister_Conflict registers a DID, then attempts to bind
// the same DID to a *different* public key; the server must reject with
// 409 DID_ALREADY_REGISTERED rather than silently overwriting.
//
// Constructing this case requires a different public key that maps to the
// same DID — which is impossible for a real did:key (the DID is the key).
// We therefore manually fabricate a request whose `did` is the canonical
// derivation of key A but whose `public_key_hex` is key B; the
// DID_KEY_MISMATCH guard fires first (400). To exercise the *registry*
// conflict path, we instead pre-poison the registry by registering key A
// and then submitting a forged request that matches key A's DID but
// supplies the same key A public key hex truncated/modified — but again
// the derivation guard fires.
//
// Net: the only way to hit the registry-level conflict is through a
// future Postgres-backed registry mutated out-of-band. The DID_KEY_MISMATCH
// case below covers the only conflict-shaped path the public API exposes
// in Phase 1, which is the security-meaningful one.
func TestIdentityRegister_DIDKeyMismatch(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	didA, _ := p256Key(t)
	_, pubHexB := p256Key(t)

	resp := ts.Post("/identity/register", "", map[string]string{
		"did":            didA,
		"public_key_hex": pubHexB,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)
	if body["code"] != "DID_KEY_MISMATCH" {
		t.Errorf("expected code DID_KEY_MISMATCH, got %v", body["code"])
	}
}

// TestIdentityRegister_MalformedBody verifies the request envelope is
// validated before any did:key parsing.
func TestIdentityRegister_MalformedBody(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	cases := []struct {
		name     string
		body     interface{}
		wantCode string
	}{
		{"missing_did", map[string]string{"public_key_hex": "04abcd"}, "MISSING_DID"},
		{"missing_pubkey", map[string]string{"did": "did:key:z2dm…"}, "MISSING_PUBLIC_KEY"},
		{"invalid_pubkey_hex", map[string]string{"did": "did:key:z2dm…", "public_key_hex": "not-hex"}, "INVALID_PUBLIC_KEY"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := ts.Post("/identity/register", "", tc.body)
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			var body map[string]interface{}
			ts.DecodeJSON(resp, &body)
			if body["code"] != tc.wantCode {
				t.Errorf("expected code %s, got %v", tc.wantCode, body["code"])
			}
		})
	}
}

// TestIdentityRegister_RawBodyMalformed posts non-JSON to confirm the
// MALFORMED_BODY branch wires correctly (the JSON-shaped MissingDID
// cases above can't reach it).
func TestIdentityRegister_RawBodyMalformed(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodPost, ts.BaseURL+"/identity/register",
		strings.NewReader("not-json{{{"))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body["code"] != "MALFORMED_BODY" {
		t.Errorf("expected code MALFORMED_BODY, got %v", body["code"])
	}
}

// TestIdentityResolve_RoundTrip registers a DID then GETs it back via
// /identity/{did}. Covers the lookup endpoint's happy path and the
// 404 path when no binding exists.
func TestIdentityResolve_RoundTrip(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	did, pubHex := p256Key(t)

	regResp := ts.Post("/identity/register", "", map[string]string{
		"did":            did,
		"public_key_hex": pubHex,
	})
	regResp.Body.Close()
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("register precondition: expected 201, got %d", regResp.StatusCode)
	}

	resp := ts.Get("/identity/"+did, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resolve: expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)
	if body["did"] != did {
		t.Errorf("resolved did = %v, want %s", body["did"], did)
	}
	if body["public_key_hex"] != pubHex {
		t.Errorf("resolved public_key_hex = %v, want %s", body["public_key_hex"], pubHex)
	}
	if body["existing"] != true {
		t.Errorf("resolved existing = %v, want true", body["existing"])
	}
}

func TestIdentityResolve_NotFound(t *testing.T) {
	ts, cleanup := testutil.StartTestServer(t)
	defer cleanup()

	unknown, _ := p256Key(t) // valid format, never registered

	resp := ts.Get("/identity/"+unknown, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	ts.DecodeJSON(resp, &body)
	if body["code"] != "DID_NOT_REGISTERED" {
		t.Errorf("expected code DID_NOT_REGISTERED, got %v", body["code"])
	}
}
