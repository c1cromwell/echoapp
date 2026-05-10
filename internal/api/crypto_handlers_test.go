package api

import (
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleServerKey_ReturnsX25519PublicKey verifies GET /v1/crypto/server-key
// returns a 200 with a valid 32-byte X25519 public key and the correct algorithm.
func TestHandleServerKey_ReturnsX25519PublicKey(t *testing.T) {
	rt := NewRouter([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/v1/crypto/server-key", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var resp ServerKeyResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Algorithm != "X25519-ChaCha20Poly1305" {
		t.Errorf("algorithm: want X25519-ChaCha20Poly1305, got %q", resp.Algorithm)
	}
	pubBytes, err := base64.StdEncoding.DecodeString(resp.PublicKey)
	if err != nil {
		t.Fatalf("public_key not valid base64: %v", err)
	}
	if len(pubBytes) != 32 {
		t.Errorf("public_key: want 32 bytes, got %d", len(pubBytes))
	}

	// Verify it's a valid X25519 public key.
	if _, err := ecdh.X25519().NewPublicKey(pubBytes); err != nil {
		t.Errorf("public_key is not a valid X25519 key: %v", err)
	}
}

// TestHandleServerKey_StableAcrossRequests ensures the same key is returned
// on repeated calls (static key, not regenerated per-request).
func TestHandleServerKey_StableAcrossRequests(t *testing.T) {
	// Reset singleton for this test — fresh router but same process-level key.
	rt := NewRouter([]string{"*"})

	getKey := func() string {
		req := httptest.NewRequest(http.MethodGet, "/v1/crypto/server-key", nil)
		w := httptest.NewRecorder()
		rt.Handler().ServeHTTP(w, req)
		var resp ServerKeyResponse
		json.NewDecoder(w.Body).Decode(&resp)
		return resp.PublicKey
	}

	k1 := getKey()
	k2 := getKey()
	if k1 != k2 {
		t.Error("server public key must be stable across requests")
	}
}

// TestHandleServerKey_MethodNotAllowed verifies POST is rejected.
func TestHandleServerKey_MethodNotAllowed(t *testing.T) {
	rt := NewRouter([]string{"*"})
	req := httptest.NewRequest(http.MethodPost, "/v1/crypto/server-key", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", w.Code)
	}
}
