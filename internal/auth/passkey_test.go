package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
)

func TestVerifyAttestation_ValidP256Key(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	// Generate a real P-256 key pair
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	// Encode public key as uncompressed (04 || x || y)
	pubBytes := elliptic.MarshalCompressed(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)
	_ = pubBytes
	pubUncompressed := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	challenge := "test-challenge-abc123"

	clientData := ClientData{
		Type:      "webauthn.create",
		Challenge: challenge,
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	attestation := AttestationResponse{
		ID:    "cred-001",
		RawID: base64.RawURLEncoding.EncodeToString([]byte("cred-001")),
		Response: AttestationResponseDetail{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AttestationObject: base64.RawURLEncoding.EncodeToString(pubUncompressed),
		},
		Type: "public-key",
	}

	pubKey, err := verifier.VerifyAttestation(attestation, challenge)
	if err != nil {
		t.Fatalf("VerifyAttestation failed: %v", err)
	}

	if len(pubKey) != 65 || pubKey[0] != 0x04 {
		t.Errorf("expected 65-byte uncompressed key, got %d bytes", len(pubKey))
	}
}

func TestVerifyAttestation_WrongChallenge(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	clientData := ClientData{
		Type:      "webauthn.create",
		Challenge: "wrong-challenge",
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	attestation := AttestationResponse{
		ID: "cred-001",
		Response: AttestationResponseDetail{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("not-a-key")),
		},
	}

	_, err := verifier.VerifyAttestation(attestation, "expected-challenge")
	if err == nil {
		t.Error("expected challenge mismatch error")
	}
}

func TestVerifyAttestation_WrongOrigin(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	clientData := ClientData{
		Type:      "webauthn.create",
		Challenge: "test",
		Origin:    "https://evil.com",
	}
	cdJSON, _ := json.Marshal(clientData)

	attestation := AttestationResponse{
		ID: "cred-001",
		Response: AttestationResponseDetail{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("key")),
		},
	}

	_, err := verifier.VerifyAttestation(attestation, "test")
	if err == nil {
		t.Error("expected origin mismatch error")
	}
}

func TestVerifyAttestation_WrongCeremonyType(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	clientData := ClientData{
		Type:      "webauthn.get", // wrong — should be webauthn.create
		Challenge: "test",
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	attestation := AttestationResponse{
		ID: "cred-001",
		Response: AttestationResponseDetail{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AttestationObject: base64.RawURLEncoding.EncodeToString([]byte("key")),
		},
	}

	_, err := verifier.VerifyAttestation(attestation, "test")
	if err == nil {
		t.Error("expected ceremony type error")
	}
}

func TestVerifyAssertion_ValidSignature(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	// Generate key pair
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubUncompressed := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	challenge := "login-challenge-xyz"

	// Build clientDataJSON
	clientData := ClientData{
		Type:      "webauthn.get",
		Challenge: challenge,
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	// Build authenticator data: rpIdHash (32) + flags (1) + signCount (4)
	rpHash := sha256.Sum256([]byte("echo.app"))
	authData := make([]byte, 37)
	copy(authData[:32], rpHash[:])
	authData[32] = 0x01 // user present flag

	// Sign: authenticatorData || SHA-256(clientDataJSON)
	cdHash := sha256.Sum256(cdJSON)
	signedData := append(authData, cdHash[:]...)
	digest := sha256.Sum256(signedData)

	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		t.Fatal(err)
	}

	// Raw r||s signature (32 bytes each for P-256)
	sig := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	credential := LoginCredential{
		ID:    "cred-001",
		RawID: base64.RawURLEncoding.EncodeToString([]byte("cred-001")),
		Response: AssertionResponseData{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AuthenticatorData: base64.RawURLEncoding.EncodeToString(authData),
			Signature:         base64.RawURLEncoding.EncodeToString(sig),
		},
		Type: "public-key",
	}

	err = verifier.VerifyAssertion(credential, pubUncompressed, challenge)
	if err != nil {
		t.Fatalf("VerifyAssertion failed: %v", err)
	}
}

func TestVerifyAssertion_InvalidSignature(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubUncompressed := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	challenge := "login-challenge-xyz"

	clientData := ClientData{
		Type:      "webauthn.get",
		Challenge: challenge,
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	rpHash := sha256.Sum256([]byte("echo.app"))
	authData := make([]byte, 37)
	copy(authData[:32], rpHash[:])
	authData[32] = 0x01

	// Wrong signature (random bytes)
	badSig := make([]byte, 64)
	rand.Read(badSig)

	credential := LoginCredential{
		ID: "cred-001",
		Response: AssertionResponseData{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AuthenticatorData: base64.RawURLEncoding.EncodeToString(authData),
			Signature:         base64.RawURLEncoding.EncodeToString(badSig),
		},
	}

	err := verifier.VerifyAssertion(credential, pubUncompressed, challenge)
	if err == nil {
		t.Error("expected signature verification failure")
	}
}

func TestVerifyAssertion_WrongRPID(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubUncompressed := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	challenge := "test"

	clientData := ClientData{
		Type:      "webauthn.get",
		Challenge: challenge,
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	// Use wrong RP ID in authenticator data
	rpHash := sha256.Sum256([]byte("evil.app"))
	authData := make([]byte, 37)
	copy(authData[:32], rpHash[:])
	authData[32] = 0x01

	credential := LoginCredential{
		ID: "cred-001",
		Response: AssertionResponseData{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AuthenticatorData: base64.RawURLEncoding.EncodeToString(authData),
			Signature:         base64.RawURLEncoding.EncodeToString(make([]byte, 64)),
		},
	}

	err := verifier.VerifyAssertion(credential, pubUncompressed, challenge)
	if err == nil {
		t.Error("expected RP ID mismatch error")
	}
}

func TestVerifyAssertion_DERSignature(t *testing.T) {
	verifier := NewPasskeyVerifier("echo.app", "https://echo.app")

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubUncompressed := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	challenge := "der-test"

	clientData := ClientData{
		Type:      "webauthn.get",
		Challenge: challenge,
		Origin:    "https://echo.app",
	}
	cdJSON, _ := json.Marshal(clientData)

	rpHash := sha256.Sum256([]byte("echo.app"))
	authData := make([]byte, 37)
	copy(authData[:32], rpHash[:])
	authData[32] = 0x01

	cdHash := sha256.Sum256(cdJSON)
	signedData := append(authData, cdHash[:]...)
	digest := sha256.Sum256(signedData)

	r, s, _ := ecdsa.Sign(rand.Reader, key, digest[:])

	// Encode as DER
	derSig := encodeDER(r, s)

	credential := LoginCredential{
		ID: "cred-001",
		Response: AssertionResponseData{
			ClientDataJSON:    base64.RawURLEncoding.EncodeToString(cdJSON),
			AuthenticatorData: base64.RawURLEncoding.EncodeToString(authData),
			Signature:         base64.RawURLEncoding.EncodeToString(derSig),
		},
	}

	err := verifier.VerifyAssertion(credential, pubUncompressed, challenge)
	if err != nil {
		t.Fatalf("DER signature verification failed: %v", err)
	}
}

func TestParseP256PublicKey_Valid(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	raw := elliptic.Marshal(elliptic.P256(), key.PublicKey.X, key.PublicKey.Y)

	parsed, err := parseP256PublicKey(raw)
	if err != nil {
		t.Fatalf("parseP256PublicKey failed: %v", err)
	}
	if parsed.X.Cmp(key.PublicKey.X) != 0 || parsed.Y.Cmp(key.PublicKey.Y) != 0 {
		t.Error("parsed key doesn't match original")
	}
}

func TestParseP256PublicKey_Invalid(t *testing.T) {
	_, err := parseP256PublicKey([]byte("too-short"))
	if err == nil {
		t.Error("expected error for invalid key")
	}

	// Valid length but not on curve
	bad := make([]byte, 65)
	bad[0] = 0x04
	_, err = parseP256PublicKey(bad)
	if err == nil {
		t.Error("expected error for point not on curve")
	}
}

// encodeDER produces a DER-encoded ECDSA signature.
func encodeDER(r, s *big.Int) []byte {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Add leading zero if high bit is set (DER integer encoding)
	if rBytes[0]&0x80 != 0 {
		rBytes = append([]byte{0x00}, rBytes...)
	}
	if sBytes[0]&0x80 != 0 {
		sBytes = append([]byte{0x00}, sBytes...)
	}

	// 0x02 || len(r) || r || 0x02 || len(s) || s
	inner := make([]byte, 0, 2+len(rBytes)+2+len(sBytes))
	inner = append(inner, 0x02, byte(len(rBytes)))
	inner = append(inner, rBytes...)
	inner = append(inner, 0x02, byte(len(sBytes)))
	inner = append(inner, sBytes...)

	// 0x30 || totalLen || inner
	der := make([]byte, 0, 2+len(inner))
	der = append(der, 0x30, byte(len(inner)))
	der = append(der, inner...)
	return der
}
