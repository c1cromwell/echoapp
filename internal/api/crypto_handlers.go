package api

// WO-13: Backend-side key exchange for X25519+ChaCha20-Poly1305 (Kinnami).
//
// The backend holds a static X25519 key pair.  iOS clients fetch the backend's
// public key from GET /v1/crypto/server-key, then use it for
// EncryptWithKeyAgreement before sending service payloads (health data, status
// queries, etc.) that the backend needs to read.
//
// Message blobs relayed between users are never decrypted by the backend —
// the content-blind relay passes them through opaquely.

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
)

// ServerKeyProvider holds the backend's static X25519 key pair.
// Generated once at startup; persisted in memory only.
type ServerKeyProvider struct {
	once    sync.Once
	privKey *ecdh.PrivateKey
	pubB64  string // base64-encoded 32-byte raw X25519 public key
}

var globalServerKeys = &ServerKeyProvider{}

// GlobalServerKeys returns the process-level X25519 key provider.
// Exposed for integration tests that need to decrypt test payloads.
func GlobalServerKeys() *ServerKeyProvider { return globalServerKeys }

// PublicKeyBase64 returns the base64-encoded X25519 public key, generating it
// lazily on first call.
func (p *ServerKeyProvider) PublicKeyBase64() (string, error) {
	var initErr error
	p.once.Do(func() {
		priv, err := ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			initErr = err
			return
		}
		p.privKey = priv
		p.pubB64 = base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes())
	})
	return p.pubB64, initErr
}

// PrivateKey returns the static private key (used by tests and service handlers
// that need to decrypt inbound X25519 payloads).
func (p *ServerKeyProvider) PrivateKey() *ecdh.PrivateKey {
	_, _ = p.PublicKeyBase64()
	return p.privKey
}

// ServerKeyResponse is the JSON payload returned by GET /v1/crypto/server-key.
type ServerKeyResponse struct {
	PublicKey string `json:"public_key"` // base64-encoded X25519 raw public key (32 bytes)
	Algorithm string `json:"algorithm"`  // always "X25519-ChaCha20Poly1305"
}

// handleServerKey serves the backend's static X25519 public key.
// GET /v1/crypto/server-key  (no auth required — public key is not secret)
func (rt *Router) handleServerKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	pubB64, err := globalServerKeys.PublicKeyBase64()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "KEY_GEN_FAILED", "failed to initialise server key", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, ServerKeyResponse{
		PublicKey: pubB64,
		Algorithm: "X25519-ChaCha20Poly1305",
	})
}
