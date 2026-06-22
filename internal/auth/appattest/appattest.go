// Package appattest implements server-side verification of Apple App Attest
// attestations and assertions, so a client's device-integrity claims (and the
// device public key bound to a DID) are cryptographically proven rather than
// trusted from a plaintext header.
//
// Production pins Apple's "App Attest Root CA" (operator-provided PEM — never
// hardcode a root from memory); tests inject a synthetic root so the full chain +
// nonce + assertion logic is unit-testable without real hardware blobs.
//
// Apple references:
//   - Attesting Key (one-time):   https://developer.apple.com/documentation/devicecheck
//   - Validating Apps That Connect to Your Server (the 8-step attestation check)
package appattest

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// nonceExtensionOID is the credCert extension carrying the attestation nonce
// (Apple-assigned OID 1.2.840.113635.100.8.2).
var nonceExtensionOID = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}

var (
	ErrBadAttestation = errors.New("appattest: invalid attestation")
	ErrBadAssertion   = errors.New("appattest: invalid assertion")
	ErrCertChain      = errors.New("appattest: certificate chain does not validate to the trusted root")
	ErrNonceMismatch  = errors.New("appattest: nonce does not match")
	ErrKeyIDMismatch  = errors.New("appattest: keyID does not match attested public key")
	ErrRPIDMismatch   = errors.New("appattest: app ID hash mismatch")
	ErrSignCount      = errors.New("appattest: sign count did not increase")
)

// Verifier validates attestations and assertions for one app (appID = teamID.bundleID)
// against a configured trust root.
type Verifier struct {
	appID    string
	roots    *x509.CertPool
	appIDSum [32]byte
}

// NewVerifier builds a verifier for appID, trusting the given root pool. In
// production pass a pool containing Apple's App Attest Root CA.
func NewVerifier(appID string, roots *x509.CertPool) *Verifier {
	return &Verifier{appID: appID, roots: roots, appIDSum: sha256.Sum256([]byte(appID))}
}

// NewVerifierFromPEM builds a verifier trusting the PEM-encoded root certificate(s).
// In production, supply Apple's "App Attest Root CA" from an operator-managed source
// (env / mounted secret) — never a certificate hardcoded from memory.
func NewVerifierFromPEM(appID string, rootPEM []byte) (*Verifier, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(rootPEM) {
		return nil, errors.New("appattest: no valid certificate found in root PEM")
	}
	return NewVerifier(appID, pool), nil
}

// attestationObject is the CBOR top-level of an App Attest attestation.
type attestationObject struct {
	Fmt      string          `cbor:"fmt"`
	AttStmt  attestationStmt `cbor:"attStmt"`
	AuthData []byte          `cbor:"authData"`
}

type attestationStmt struct {
	X5C     [][]byte `cbor:"x5c"`
	Receipt []byte   `cbor:"receipt"`
}

// AttestationResult is the outcome of a successful attestation.
type AttestationResult struct {
	PublicKey *ecdsa.PublicKey
	SignCount uint32
}

// VerifyAttestation runs Apple's attestation checks and returns the attested public
// key (to be stored against keyID for later assertion verification).
//
//	attestation: the CBOR attestation object from DCAppAttestService.attestKey
//	challenge:   the server-issued one-time challenge the client attested over
//	keyID:       base64-decoded key identifier (== SHA256 of the public key)
func (v *Verifier) VerifyAttestation(attestation, challenge, keyID []byte) (*AttestationResult, error) {
	var obj attestationObject
	if err := cbor.Unmarshal(attestation, &obj); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrBadAttestation, err)
	}
	if obj.Fmt != "apple-appattest" {
		return nil, fmt.Errorf("%w: unexpected fmt %q", ErrBadAttestation, obj.Fmt)
	}
	if len(obj.AttStmt.X5C) < 2 {
		return nil, fmt.Errorf("%w: x5c must contain credCert + intermediate", ErrBadAttestation)
	}

	// 1–2. Parse + chain-verify credCert → intermediate → trusted root.
	credCert, err := x509.ParseCertificate(obj.AttStmt.X5C[0])
	if err != nil {
		return nil, fmt.Errorf("%w: parse credCert: %v", ErrBadAttestation, err)
	}
	intermediates := x509.NewCertPool()
	for _, der := range obj.AttStmt.X5C[1:] {
		ic, perr := x509.ParseCertificate(der)
		if perr != nil {
			return nil, fmt.Errorf("%w: parse intermediate: %v", ErrBadAttestation, perr)
		}
		intermediates.AddCert(ic)
	}
	if _, err := credCert.Verify(x509.VerifyOptions{
		Roots:         v.roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCertChain, err)
	}

	// 3–4. nonce = SHA256(authData || SHA256(challenge)); must equal the value in the
	// credCert's Apple nonce extension (wrapped in a DER SEQUENCE > [0] > OCTET STRING).
	clientDataHash := sha256.Sum256(challenge)
	nonce := sha256.Sum256(append(append([]byte{}, obj.AuthData...), clientDataHash[:]...))
	certNonce, err := nonceFromCert(credCert)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadAttestation, err)
	}
	if string(certNonce) != string(nonce[:]) {
		return nil, ErrNonceMismatch
	}

	// 5. The attested public key is the credCert's leaf key; keyID == SHA256(pubkey)
	// in ANSI X9.63 uncompressed form.
	ecPub, ok := credCert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: credCert key is not P-256", ErrBadAttestation)
	}
	pubX963 := elliptic963(ecPub)
	pubSum := sha256.Sum256(pubX963)
	if string(pubSum[:]) != string(keyID) {
		return nil, ErrKeyIDMismatch
	}

	// 6. authData: rpIDHash(appID) + counter (+ attestedCredentialData).
	ad, err := parseAuthData(obj.AuthData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadAttestation, err)
	}
	if ad.rpIDHash != v.appIDSum {
		return nil, ErrRPIDMismatch
	}

	return &AttestationResult{PublicKey: ecPub, SignCount: ad.signCount}, nil
}

// VerifyAssertion validates a per-request assertion against the stored attested key,
// enforcing the monotonic sign count. Returns the new sign count to persist.
//
//	assertion:    CBOR assertion from DCAppAttestService.generateAssertion
//	clientData:   the exact request payload the client signed over
//	pub:          the public key stored at attestation time
//	lastSignCount: the last persisted sign count for this key
func (v *Verifier) VerifyAssertion(assertion, clientData []byte, pub *ecdsa.PublicKey, lastSignCount uint32) (newSignCount uint32, err error) {
	var a struct {
		Signature []byte `cbor:"signature"`
		AuthData  []byte `cbor:"authenticatorData"`
	}
	if err := cbor.Unmarshal(assertion, &a); err != nil {
		return 0, fmt.Errorf("%w: decode: %v", ErrBadAssertion, err)
	}

	// nonce = SHA256(authenticatorData || SHA256(clientData)); signature is ECDSA(pub, nonce).
	clientHash := sha256.Sum256(clientData)
	nonce := sha256.Sum256(append(append([]byte{}, a.AuthData...), clientHash[:]...))
	if !ecdsa.VerifyASN1(pub, nonce[:], a.Signature) {
		return 0, fmt.Errorf("%w: signature", ErrBadAssertion)
	}

	ad, err := parseAuthData(a.AuthData)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrBadAssertion, err)
	}
	if ad.rpIDHash != v.appIDSum {
		return 0, ErrRPIDMismatch
	}
	// Replay/clone protection: the counter must strictly increase.
	if ad.signCount <= lastSignCount {
		return 0, ErrSignCount
	}
	return ad.signCount, nil
}

// authData is the parsed WebAuthn-style authenticator data (App Attest reuses it).
type authData struct {
	rpIDHash  [32]byte
	flags     byte
	signCount uint32
}

func parseAuthData(b []byte) (authData, error) {
	if len(b) < 37 {
		return authData{}, errors.New("authData too short")
	}
	var ad authData
	copy(ad.rpIDHash[:], b[0:32])
	ad.flags = b[32]
	ad.signCount = uint32(b[33])<<24 | uint32(b[34])<<16 | uint32(b[35])<<8 | uint32(b[36])
	return ad, nil
}

// nonceFromCert extracts the attestation nonce from the Apple credCert extension.
// The extension value is DER: SEQUENCE { [0] EXPLICIT OCTET STRING nonce }.
func nonceFromCert(cert *x509.Certificate) ([]byte, error) {
	for _, ext := range cert.Extensions {
		if !ext.Id.Equal(nonceExtensionOID) {
			continue
		}
		var seq struct {
			Nonce []byte `asn1:"tag:0,explicit"`
		}
		if _, err := asn1.Unmarshal(ext.Value, &seq); err != nil {
			return nil, fmt.Errorf("decode nonce extension: %w", err)
		}
		return seq.Nonce, nil
	}
	return nil, errors.New("nonce extension not present in credCert")
}

// elliptic963 encodes a P-256 public key in ANSI X9.63 uncompressed form (0x04 || X || Y).
func elliptic963(pub *ecdsa.PublicKey) []byte {
	out := make([]byte, 1+32+32)
	out[0] = 0x04
	pub.X.FillBytes(out[1:33])
	pub.Y.FillBytes(out[33:65])
	return out
}
