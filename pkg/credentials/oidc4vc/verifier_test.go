package oidc4vc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thechadcromwell/echoapp/pkg/credentials"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// --- test helpers ---

func generateTestKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

// signVPJWT builds a minimal VP JWT signed with priv.
func signVPJWT(t *testing.T, priv *ecdsa.PrivateKey, holderDID string, credentials []string, expOffset time.Duration) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256","typ":"JWT"}`))

	exp := time.Now().Add(expOffset).Unix()
	if expOffset == 0 {
		exp = 0
	}
	payload := map[string]interface{}{
		"iss":   holderDID,
		"aud":   "did:key:verifier",
		"iat":   time.Now().Unix(),
		"nonce": "test-nonce",
		"vp": map[string]interface{}{
			"@context":             []string{"https://www.w3.org/2018/credentials/v1"},
			"type":                 []string{"VerifiablePresentation"},
			"verifiableCredential": credentials,
		},
	}
	if exp != 0 {
		payload["exp"] = exp
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := header + "." + payloadB64
	digest := sha256.Sum256([]byte(signingInput))

	r, s, err := ecdsa.Sign(rand.Reader, priv, digest[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	// ES256 JWT: P1363 format (R || S, 64 bytes)
	sig := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// holderDIDFromKey derives a did:key from a P-256 public key.
func holderDIDFromKey(t *testing.T, priv *ecdsa.PrivateKey) string {
	t.Helper()
	xBytes := make([]byte, 32)
	yBytes := make([]byte, 32)
	copy(xBytes[32-len(priv.PublicKey.X.Bytes()):], priv.PublicKey.X.Bytes())
	copy(yBytes[32-len(priv.PublicKey.Y.Bytes()):], priv.PublicKey.Y.Bytes())
	uncompressed := append([]byte{0x04}, append(xBytes, yBytes...)...)
	encoded := make([]byte, len(uncompressed)*2)
	_ = encoded
	did, err := didkey.DeriveFromPublicKeyHex(hexBytes(append([]byte{0x04}, append(xBytes, yBytes...)...)))
	if err != nil {
		t.Fatalf("derive DID: %v", err)
	}
	return did
}

func hexBytes(b []byte) string {
	const hextable = "0123456789abcdef"
	buf := make([]byte, len(b)*2)
	for i, v := range b {
		buf[i*2] = hextable[v>>4]
		buf[i*2+1] = hextable[v&0xf]
	}
	return string(buf)
}

// --- mock credential verifier ---

type mockCredVerifier struct {
	result *credentials.CredentialVerificationResult
	err    error
}

func (m *mockCredVerifier) VerifyCredential(_ context.Context, _ *credentials.CredentialVerificationRequest) (*credentials.CredentialVerificationResult, error) {
	return m.result, m.err
}

// --- parseVPJWT tests ---

func TestParseVPJWT_Valid(t *testing.T) {
	priv := generateTestKey(t)
	did := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, did, []string{"eyJcred1"}, time.Hour)

	vp, headerPayload, sig, err := parseVPJWT(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vp.Iss != did {
		t.Errorf("iss = %q, want %q", vp.Iss, did)
	}
	if len(vp.VP.VerifiableCredential) != 1 {
		t.Errorf("vc count = %d, want 1", len(vp.VP.VerifiableCredential))
	}
	if !strings.Contains(headerPayload, ".") {
		t.Error("rawHeaderPayload missing dot separator")
	}
	if len(sig) == 0 {
		t.Error("sig is empty")
	}
}

func TestParseVPJWT_InvalidParts(t *testing.T) {
	_, _, _, err := parseVPJWT("not.a.valid.jwt.extra")
	if err == nil {
		t.Fatal("expected error for 5-part token")
	}
}

func TestParseVPJWT_MissingVPClaim(t *testing.T) {
	// Build a JWT payload without the "vp" claim
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"did:key:z123"}`))
	_, _, _, err := parseVPJWT(header + "." + payload + ".fakesig")
	if err == nil {
		t.Fatal("expected error for missing vp claim")
	}
}

// --- VerifyPresentation tests ---

func TestVerifyPresentation_ValidVP(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"{\"@context\":[],\"type\":[\"VerifiableCredential\"]}"}, time.Hour)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	v.SetCredentialService(&mockCredVerifier{
		result: &credentials.CredentialVerificationResult{
			CredentialID: "cred-001",
			IsValid:      true,
		},
	})

	sub := &PresentationSubmission{
		ID:           "sub-001",
		DefinitionID: "echo_proof_of_humanity_v1",
		DescriptorMap: []DescriptorMap{
			{ID: "proof_of_humanity", Format: "json-ld", Path: "$"},
		},
	}

	result, err := v.VerifyPresentation(context.Background(), token, sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid {
		t.Errorf("expected valid=true, got false; error: %s", result.Error)
	}
	if result.HolderDID != holderDID {
		t.Errorf("holderDID = %q, want %q", result.HolderDID, holderDID)
	}
	if len(result.Credentials) != 1 {
		t.Errorf("credentials count = %d, want 1", len(result.Credentials))
	}
}

func TestVerifyPresentation_ExpiredVP(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	// Build a VP with exp already in the past
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256","typ":"JWT"}`))
	payload := map[string]interface{}{
		"iss": holderDID,
		"aud": "did:key:verifier",
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
		"exp": time.Now().Add(-time.Hour).Unix(), // expired
		"vp": map[string]interface{}{
			"@context":             []string{"https://www.w3.org/2018/credentials/v1"},
			"type":                 []string{"VerifiablePresentation"},
			"verifiableCredential": []string{},
		},
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payloadB64
	digest := sha256.Sum256([]byte(signingInput))
	r, s, _ := ecdsa.Sign(rand.Reader, priv, digest[:])
	sig := make([]byte, 64)
	copy(sig[32-len(r.Bytes()):32], r.Bytes())
	copy(sig[64-len(s.Bytes()):64], s.Bytes())
	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	result, err := v.VerifyPresentation(context.Background(), token, &PresentationSubmission{ID: "x", DefinitionID: "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsValid {
		t.Error("expected expired VP to be invalid")
	}
	if !strings.Contains(result.Error, "expired") {
		t.Errorf("error should mention expiry, got: %s", result.Error)
	}
}

func TestVerifyPresentation_TamperedSignature(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"cred"}, time.Hour)

	// Flip the last byte of the signature
	parts := strings.Split(token, ".")
	rawSig, _ := base64.RawURLEncoding.DecodeString(parts[2])
	rawSig[len(rawSig)-1] ^= 0xFF
	parts[2] = base64.RawURLEncoding.EncodeToString(rawSig)
	tamperedToken := strings.Join(parts, ".")

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	result, err := v.VerifyPresentation(context.Background(), tamperedToken, &PresentationSubmission{ID: "x", DefinitionID: "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsValid {
		t.Error("expected tampered VP to be invalid")
	}
	if !strings.Contains(result.Error, "signature") {
		t.Errorf("error should mention signature, got: %s", result.Error)
	}
}

func TestVerifyPresentation_MissingVPToken(t *testing.T) {
	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	_, err := v.VerifyPresentation(context.Background(), "", &PresentationSubmission{})
	if err == nil {
		t.Fatal("expected error for empty vp_token")
	}
}

func TestVerifyPresentation_MissingSubmission(t *testing.T) {
	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	_, err := v.VerifyPresentation(context.Background(), "a.b.c", nil)
	if err == nil {
		t.Fatal("expected error for nil submission")
	}
}

func TestVerifyPresentation_NoCredentialVerifier(t *testing.T) {
	// When no credential verifier is wired, individual VCs are marked valid by default.
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"cred1", "cred2"}, time.Hour)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	result, err := v.VerifyPresentation(context.Background(), token, &PresentationSubmission{ID: "x", DefinitionID: "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid {
		t.Errorf("expected valid=true when no verifier wired; error: %s", result.Error)
	}
}

func TestVerifyPresentation_CredentialVerifierFails(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"bad-cred"}, time.Hour)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	v.SetCredentialService(&mockCredVerifier{
		result: &credentials.CredentialVerificationResult{
			IsValid: false,
			Errors: []credentials.VerificationError{
				{Code: "INVALID_SIGNATURE", Message: "bad proof"},
			},
		},
	})

	result, err := v.VerifyPresentation(context.Background(), token, &PresentationSubmission{ID: "x", DefinitionID: "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsValid {
		t.Error("expected invalid when credential fails")
	}
	if len(result.Credentials) == 0 || result.Credentials[0].IsValid {
		t.Error("expected first credential entry to be invalid")
	}
}

// --- HTTP handler tests ---

func setupTestRouter(v *Verifier) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v.RegisterRoutes(r)
	return r
}

func TestSubmitPresentation_HTTP_Valid(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"cred1"}, time.Hour)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	v.storeChallenge("test-state-001") // S8: state must be a server-issued challenge
	r := setupTestRouter(v)

	body, _ := json.Marshal(map[string]interface{}{
		"vp_token": token,
		"state":    "test-state-001",
		"presentation_submission": map[string]interface{}{
			"id":            "sub-001",
			"definition_id": "echo_proof_of_humanity_v1",
			"descriptor_map": []map[string]interface{}{
				{"id": "proof_of_humanity", "format": "jwt_vc_json", "path": "$"},
			},
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verification/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// S8: replaying the same submission must fail — the challenge is single-use.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/verification/submit", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("replayed submission want 400, got %d", rec2.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["presentationId"] == "" {
		t.Error("missing presentationId in response")
	}
}

func TestSubmitPresentation_HTTP_MissingVPToken(t *testing.T) {
	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	r := setupTestRouter(v)

	body, _ := json.Marshal(map[string]interface{}{
		"state": "s",
		"presentation_submission": map[string]interface{}{
			"id": "x", "definition_id": "y", "descriptor_map": []interface{}{},
		},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verification/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// TestSubmitPresentation_UnknownState verifies S8: a state the verifier never
// issued is rejected, blocking forged or out-of-band submissions.
func TestSubmitPresentation_UnknownState(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"c1"}, time.Hour)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	r := setupTestRouter(v) // no challenge stored

	body, _ := json.Marshal(map[string]interface{}{
		"vp_token": token,
		"state":    "never-issued",
		"presentation_submission": map[string]interface{}{
			"id": "x", "definition_id": "y", "descriptor_map": []interface{}{},
		},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/verification/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown state want 400, got %d", rec.Code)
	}
}

func TestGetVerificationStatus_NotFound(t *testing.T) {
	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	r := setupTestRouter(v)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/verification/pres_nonexistent/status", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestGetVerificationStatus_Found(t *testing.T) {
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)
	token := signVPJWT(t, priv, holderDID, []string{"c1"}, time.Hour)

	v := NewVerifier("did:key:verifier", "did:key:issuer", "http://localhost", "http://localhost")
	v.storeChallenge("mystate") // S8: state must be a server-issued challenge
	r := setupTestRouter(v)

	// Submit first
	body, _ := json.Marshal(map[string]interface{}{
		"vp_token": token,
		"state":    "mystate",
		"presentation_submission": map[string]interface{}{
			"id":             "sub",
			"definition_id":  "echo_kyc_lite_v1",
			"descriptor_map": []interface{}{},
		},
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/verification/submit",
		bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("submit: want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Now fetch status
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/verification/pres_mystate/status", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

// --- detectCredentialFormat tests ---

func TestDetectCredentialFormat(t *testing.T) {
	cases := []struct {
		token  string
		format credentials.CredentialFormat
	}{
		{`{"@context":[], "type":[]}`, credentials.JSONLDFormat},
		{"eyJhbGci.payload.sig", credentials.JWTFormat},
		{"eyJhbGci.payload.sig~disclosure", credentials.SDJWTFormat},
	}
	for _, tc := range cases {
		got := detectCredentialFormat(tc.token)
		if got != tc.format {
			t.Errorf("detectCredentialFormat(%q) = %q, want %q", tc.token[:min(20, len(tc.token))], got, tc.format)
		}
	}
}

// --- parseCredentialTypeFromDefinitionID tests ---

func TestParseCredentialTypeFromDefinitionID(t *testing.T) {
	cases := []struct{ id, want string }{
		{"echo_proof_of_humanity_v1", "ProofOfHumanity"},
		{"echo_kyc_lite_v1", "KYCLite"},
		{"echo_unknown_v1", "ProofOfHumanity"},
		{"", "ProofOfHumanity"},
	}
	for _, tc := range cases {
		got := parseCredentialTypeFromDefinitionID(tc.id)
		if got != tc.want {
			t.Errorf("%q → %q, want %q", tc.id, got, tc.want)
		}
	}
}

// --- sign with ASN.1 DER for verifyVPHolderSignature test ---

func TestVerifyVPHolderSignature_ASN1(t *testing.T) {
	// Verify that verifyVPHolderSignature also works with ASN.1 DER-encoded signatures
	// (the VerifyECDSAP256SHA256 function accepts both P1363 and ASN.1 DER).
	priv := generateTestKey(t)
	holderDID := holderDIDFromKey(t, priv)

	msg := []byte("header.payload")
	digest := sha256.Sum256(msg)
	r, s, _ := ecdsa.Sign(rand.Reader, priv, digest[:])
	derSig, _ := asn1.Marshal(struct{ R, S *big.Int }{r, s})

	if err := verifyVPHolderSignature("header.payload", derSig, holderDID); err != nil {
		t.Errorf("ASN.1 sig should verify: %v", err)
	}
}
