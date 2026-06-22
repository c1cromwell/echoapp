package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
)

// TestPublicJWKS_VerifiesIssuedToken proves the published JWK Set advertises the
// real signing key: reconstruct the public key from the JWK's x/y coordinates and
// confirm it validates the ES256 signature of an actually-issued token, with a
// matching kid.
func TestPublicJWKS_VerifiesIssuedToken(t *testing.T) {
	ts := newTestTokenService(t)

	token, _, err := ts.IssueAccessToken("did:key:z6MkJWKSubjectXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "device-1", 1, "messaging")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}

	// kid from the token header.
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	if header.Alg != "ES256" {
		t.Fatalf("expected ES256, got %s", header.Alg)
	}

	// Parse JWKS.
	jwksBytes, err := ts.PublicJWKS()
	if err != nil {
		t.Fatalf("PublicJWKS: %v", err)
	}
	var jwks struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(jwksBytes, &jwks); err != nil {
		t.Fatalf("unmarshal jwks: %v", err)
	}
	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(jwks.Keys))
	}
	jwk := jwks.Keys[0]
	if jwk["kty"] != "EC" || jwk["crv"] != "P-256" || jwk["alg"] != "ES256" {
		t.Fatalf("unexpected JWK params: %+v", jwk)
	}
	if jwk["kid"] != header.Kid {
		t.Fatalf("kid mismatch: token=%s jwks=%s", header.Kid, jwk["kid"])
	}

	// Reconstruct the public key from x/y and verify the token signature.
	xb, err := base64.RawURLEncoding.DecodeString(jwk["x"])
	if err != nil || len(xb) != 32 {
		t.Fatalf("bad x coord (len=%d, err=%v)", len(xb), err)
	}
	yb, err := base64.RawURLEncoding.DecodeString(jwk["y"])
	if err != nil || len(yb) != 32 {
		t.Fatalf("bad y coord (len=%d, err=%v)", len(yb), err)
	}
	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xb),
		Y:     new(big.Int).SetBytes(yb),
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) != 64 {
		t.Fatalf("bad signature (len=%d, err=%v)", len(sig), err)
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(pub, hash[:], r, s) {
		t.Fatal("JWKS public key failed to verify an issued token signature")
	}
}
