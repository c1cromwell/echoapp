package api

// Golden-vector tests for WO-1 ECDSA P-256 passkey auth middleware.
//
// Vectors are pre-generated deterministic inputs (fixed private key + fixed body).
// The signature itself is produced fresh each run from the fixed key — Go's ECDSA
// Sign uses randomness but the key and body are the deterministic inputs per WO-1 spec.

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Static test vectors ---
//
// Generated with: go run (see test/crypto_vectors/)
// Private key d: eaff1084e1322774ce79ad393aa9d925d73473a9de5fe6a50d969672ac66be4f
// Public key (uncompressed 04||X||Y):
//   047bfc587ef5617b74f66c8c26765adfc6ac311be92ec5546f146b28a026e96ed
//   d56e82054fa9de5114d4e59938f29dbe2f63bffa40a321f0f472f28ea30f8d1f4

const (
	vectorPrivKeyHex = "eaff1084e1322774ce79ad393aa9d925d73473a9de5fe6a50d969672ac66be4f"
	vectorPubKeyHex  = "047bfc587ef5617b74f66c8c26765adfc6ac311be92ec5546f146b28a026e96edd56e82054fa9de5114d4e59938f29dbe2f63bffa40a321f0f472f28ea30f8d1f4"
	vectorBody       = `{"hello":"phase1"}`
	vectorDID        = "did:key:test-vector-passkey"
)

// vectorSign produces an ECDSA P-256 ASN.1 DER signature over SHA-256(body) using
// the fixed private key, then returns it as standard base64.
func vectorSign(t *testing.T, body []byte) string {
	t.Helper()
	privD := mustHexBigInt(t, vectorPrivKeyHex)
	curve := elliptic.P256()
	priv := &ecdsa.PrivateKey{
		D: privD,
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     mustHexBigInt(t, vectorPubKeyHex[2:66]),
			Y:     mustHexBigInt(t, vectorPubKeyHex[66:]),
		},
	}
	digest := sha256.Sum256(body)
	r, s, err := ecdsa.Sign(rand.Reader, priv, digest[:])
	if err != nil {
		t.Fatalf("ecdsa.Sign: %v", err)
	}
	der, err := asn1.Marshal(struct{ R, S *big.Int }{r, s})
	if err != nil {
		t.Fatalf("asn1.Marshal: %v", err)
	}
	return base64.StdEncoding.EncodeToString(der)
}

func mustHexBigInt(t *testing.T, h string) *big.Int {
	t.Helper()
	b, err := hex.DecodeString(h)
	if err != nil {
		t.Fatalf("hex decode %q: %v", h, err)
	}
	return new(big.Int).SetBytes(b)
}

// --- Minimal DID registry for tests ---

type singleKeyRegistry struct{ hexKey string }

func (r *singleKeyRegistry) Register(_ context.Context, _, _ string) (*DIDBinding, bool, error) {
	return nil, false, ErrDIDConflict
}
func (r *singleKeyRegistry) Lookup(_ context.Context, did string) (*DIDBinding, error) {
	return &DIDBinding{DID: did, PublicKeyHex: r.hexKey}, nil
}
func (r *singleKeyRegistry) ListDevices(_ context.Context, did string) ([]*DIDBinding, error) {
	return []*DIDBinding{{DID: did, PublicKeyHex: r.hexKey}}, nil
}
func (r *singleKeyRegistry) RegisterAdditionalDevice(_ context.Context, _, _, _ string) (*DIDBinding, error) {
	return nil, nil
}

// emptyRegistry returns ErrBindingNotFound for every DID.
type emptyRegistry struct{}

func (r *emptyRegistry) Register(_ context.Context, _, _ string) (*DIDBinding, bool, error) {
	return nil, false, ErrDIDConflict
}
func (r *emptyRegistry) Lookup(_ context.Context, _ string) (*DIDBinding, error) {
	return nil, ErrBindingNotFound
}
func (r *emptyRegistry) ListDevices(_ context.Context, _ string) ([]*DIDBinding, error) {
	return nil, ErrBindingNotFound
}
func (r *emptyRegistry) RegisterAdditionalDevice(_ context.Context, _, _, _ string) (*DIDBinding, error) {
	return nil, ErrBindingNotFound
}

func buildTestRouter(reg DIDRegistry) *Router {
	return &Router{
		DIDRegistry:     reg,
		TokenValidator:  func(string) bool { return false },
		UserIDExtractor: func(string) string { return "" },
	}
}

func passkeyRequest(did, sig, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/v1/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if did != "" {
		req.Header.Set(headerSenderDID, did)
	}
	if sig != "" {
		req.Header.Set(headerSignature, sig)
	}
	return req
}

// TestAuthMiddleware_ValidSignature — known key + known body + fresh sig → 200, DID in context
func TestAuthMiddleware_ValidSignature(t *testing.T) {
	rt := buildTestRouter(&singleKeyRegistry{hexKey: vectorPubKeyHex})
	sig := vectorSign(t, []byte(vectorBody))

	var capturedDID string
	handler := rt.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if did, ok := r.Context().Value(ContextKeyUserID).(string); ok {
			capturedDID = did
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, passkeyRequest(vectorDID, sig, vectorBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if capturedDID != vectorDID {
		t.Errorf("context DID = %q, want %q", capturedDID, vectorDID)
	}
}

// TestAuthMiddleware_TamperedSignature — flip last byte of valid sig → 401 AUTH_INVALID_SIGNATURE
func TestAuthMiddleware_TamperedSignature(t *testing.T) {
	rt := buildTestRouter(&singleKeyRegistry{hexKey: vectorPubKeyHex})
	sig := vectorSign(t, []byte(vectorBody))

	raw, _ := base64.StdEncoding.DecodeString(sig)
	raw[len(raw)-1] ^= 0xFF
	tamperedSig := base64.StdEncoding.EncodeToString(raw)

	handler := rt.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, passkeyRequest(vectorDID, tamperedSig, vectorBody))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("AUTH_INVALID_SIGNATURE")) {
		t.Errorf("body missing AUTH_INVALID_SIGNATURE: %s", rec.Body.String())
	}
}

// TestAuthMiddleware_MissingHeader — X-Sender-DID present, X-Signature absent → 401 AUTH_MISSING_SIGNATURE
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	rt := buildTestRouter(&singleKeyRegistry{hexKey: vectorPubKeyHex})

	handler := rt.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := passkeyRequest(vectorDID, "" /* no sig */, vectorBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("AUTH_MISSING_SIGNATURE")) {
		t.Errorf("body missing AUTH_MISSING_SIGNATURE: %s", rec.Body.String())
	}
}

// TestAuthMiddleware_UnknownDID — DID not in registry → 401 AUTH_UNKNOWN_DID
func TestAuthMiddleware_UnknownDID(t *testing.T) {
	rt := buildTestRouter(&emptyRegistry{})
	sig := vectorSign(t, []byte(vectorBody))

	handler := rt.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, passkeyRequest("did:key:not-registered", sig, vectorBody))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("AUTH_UNKNOWN_DID")) {
		t.Errorf("body missing AUTH_UNKNOWN_DID: %s", rec.Body.String())
	}
}
