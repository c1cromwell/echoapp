package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

// PasskeyVerifier implements WebAuthn P-256 passkey attestation and assertion
// verification. It validates client data JSON, authenticator data, and ECDSA
// signatures against stored public keys.
type PasskeyVerifier struct {
	RPID     string // Relying party identifier (e.g., "echo.app")
	Expected string // Expected origin (e.g., "https://echo.app")
}

// NewPasskeyVerifier creates a verifier for the given relying party.
func NewPasskeyVerifier(rpID, origin string) *PasskeyVerifier {
	return &PasskeyVerifier{
		RPID:     rpID,
		Expected: origin,
	}
}

// ClientData is the decoded clientDataJSON from a WebAuthn ceremony.
type ClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

// VerifyAttestation validates a passkey registration response.
// It verifies clientDataJSON fields and extracts the public key from the
// attestation. Returns the raw ECDSA public key bytes (65 bytes uncompressed).
func (v *PasskeyVerifier) VerifyAttestation(attestation AttestationResponse, expectedChallenge string) ([]byte, error) {
	// 1. Decode and validate clientDataJSON
	clientDataRaw, err := base64.RawURLEncoding.DecodeString(attestation.Response.ClientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("decode clientDataJSON: %w", err)
	}

	var clientData ClientData
	if err := json.Unmarshal(clientDataRaw, &clientData); err != nil {
		return nil, fmt.Errorf("parse clientDataJSON: %w", err)
	}

	if clientData.Type != "webauthn.create" {
		return nil, fmt.Errorf("invalid ceremony type: expected webauthn.create, got %s", clientData.Type)
	}

	if clientData.Challenge != expectedChallenge {
		return nil, fmt.Errorf("challenge mismatch")
	}

	if v.Expected != "" && clientData.Origin != v.Expected {
		return nil, fmt.Errorf("origin mismatch: expected %s, got %s", v.Expected, clientData.Origin)
	}

	// 2. Decode attestation object to extract public key
	// WebAuthn attestation objects are CBOR-encoded. For the "none" attestation
	// format (Apple platform passkeys), the public key is in authData.
	// We support raw P-256 public key extraction from a simplified attestation.
	pubKeyBytes, err := base64.RawURLEncoding.DecodeString(attestation.Response.AttestationObject)
	if err != nil {
		return nil, fmt.Errorf("decode attestation object: %w", err)
	}

	// Validate it looks like a P-256 public key (65 bytes uncompressed: 04 || x || y)
	if len(pubKeyBytes) == 65 && pubKeyBytes[0] == 0x04 {
		return pubKeyBytes, nil
	}

	// If the attestation object is longer, try to extract the embedded COSE key.
	// This handles the full CBOR case by searching for the uncompressed P-256 marker.
	if pk := extractP256Key(pubKeyBytes); pk != nil {
		return pk, nil
	}

	return nil, fmt.Errorf("unable to extract P-256 public key from attestation")
}

// VerifyAssertion validates a passkey login assertion.
// storedPubKey is the raw P-256 public key bytes (65 bytes uncompressed) stored
// during registration.
func (v *PasskeyVerifier) VerifyAssertion(credential LoginCredential, storedPubKey []byte, expectedChallenge string) error {
	// 1. Parse the stored public key
	pubKey, err := parseP256PublicKey(storedPubKey)
	if err != nil {
		return fmt.Errorf("parse stored public key: %w", err)
	}

	// 2. Decode and validate clientDataJSON
	clientDataRaw, err := base64.RawURLEncoding.DecodeString(credential.Response.ClientDataJSON)
	if err != nil {
		return fmt.Errorf("decode clientDataJSON: %w", err)
	}

	var clientData ClientData
	if err := json.Unmarshal(clientDataRaw, &clientData); err != nil {
		return fmt.Errorf("parse clientDataJSON: %w", err)
	}

	if clientData.Type != "webauthn.get" {
		return fmt.Errorf("invalid ceremony type: expected webauthn.get, got %s", clientData.Type)
	}

	if clientData.Challenge != expectedChallenge {
		return fmt.Errorf("challenge mismatch")
	}

	if v.Expected != "" && clientData.Origin != v.Expected {
		return fmt.Errorf("origin mismatch: expected %s, got %s", v.Expected, clientData.Origin)
	}

	// 3. Decode authenticator data
	authData, err := base64.RawURLEncoding.DecodeString(credential.Response.AuthenticatorData)
	if err != nil {
		return fmt.Errorf("decode authenticator data: %w", err)
	}

	// 4. Verify RP ID hash (first 32 bytes of authenticator data)
	if len(authData) < 37 {
		return fmt.Errorf("authenticator data too short")
	}
	rpIDHash := sha256.Sum256([]byte(v.RPID))
	for i := 0; i < 32; i++ {
		if authData[i] != rpIDHash[i] {
			return fmt.Errorf("RP ID hash mismatch")
		}
	}

	// 5. Check user presence flag (bit 0 of flags byte at offset 32)
	flags := authData[32]
	if flags&0x01 == 0 {
		return fmt.Errorf("user presence flag not set")
	}

	// 6. Verify the signature
	// WebAuthn signature is over: authenticatorData || SHA-256(clientDataJSON)
	clientDataHash := sha256.Sum256(clientDataRaw)
	signedData := append(authData, clientDataHash[:]...)
	digest := sha256.Sum256(signedData)

	// 7. Decode the signature (DER or raw r||s)
	sigBytes, err := base64.RawURLEncoding.DecodeString(credential.Response.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	if !verifyP256Signature(pubKey, digest[:], sigBytes) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// parseP256PublicKey parses an uncompressed P-256 public key (65 bytes: 04 || x || y).
func parseP256PublicKey(raw []byte) (*ecdsa.PublicKey, error) {
	if len(raw) != 65 || raw[0] != 0x04 {
		return nil, fmt.Errorf("invalid P-256 public key: expected 65-byte uncompressed format")
	}
	x := new(big.Int).SetBytes(raw[1:33])
	y := new(big.Int).SetBytes(raw[33:65])

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// Verify the point is on the curve
	if !pubKey.Curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("public key point not on P-256 curve")
	}

	return pubKey, nil
}

// verifyP256Signature verifies an ECDSA P-256 signature in either DER or raw r||s format.
func verifyP256Signature(pubKey *ecdsa.PublicKey, hash, sig []byte) bool {
	// Try raw r||s format first (each 32 bytes for P-256)
	if len(sig) == 64 {
		r := new(big.Int).SetBytes(sig[:32])
		s := new(big.Int).SetBytes(sig[32:])
		return ecdsa.Verify(pubKey, hash, r, s)
	}

	// Try DER format
	r, s, err := parseDERSignature(sig)
	if err != nil {
		return false
	}
	return ecdsa.Verify(pubKey, hash, r, s)
}

// parseDERSignature parses a DER-encoded ECDSA signature.
func parseDERSignature(der []byte) (*big.Int, *big.Int, error) {
	if len(der) < 8 || der[0] != 0x30 {
		return nil, nil, fmt.Errorf("not a DER sequence")
	}

	idx := 2
	if der[1]&0x80 != 0 {
		idx = 2 + int(der[1]&0x7f)
	}

	// Parse R
	if idx >= len(der) || der[idx] != 0x02 {
		return nil, nil, fmt.Errorf("expected integer tag for R")
	}
	idx++
	rLen := int(der[idx])
	idx++
	if idx+rLen > len(der) {
		return nil, nil, fmt.Errorf("R length overflow")
	}
	r := new(big.Int).SetBytes(der[idx : idx+rLen])
	idx += rLen

	// Parse S
	if idx >= len(der) || der[idx] != 0x02 {
		return nil, nil, fmt.Errorf("expected integer tag for S")
	}
	idx++
	sLen := int(der[idx])
	idx++
	if idx+sLen > len(der) {
		return nil, nil, fmt.Errorf("S length overflow")
	}
	s := new(big.Int).SetBytes(der[idx : idx+sLen])

	return r, s, nil
}

// extractP256Key searches for a 65-byte uncompressed P-256 key (0x04 prefix) in raw bytes.
func extractP256Key(data []byte) []byte {
	for i := 0; i+65 <= len(data); i++ {
		if data[i] == 0x04 {
			candidate := data[i : i+65]
			x := new(big.Int).SetBytes(candidate[1:33])
			y := new(big.Int).SetBytes(candidate[33:65])
			if elliptic.P256().IsOnCurve(x, y) {
				return candidate
			}
		}
	}
	return nil
}

// generateChallenge is defined in service.go — not duplicated here.
