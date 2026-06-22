package appattest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
)

const testAppID = "ABCDE12345.com.echo.app"

// testPKI builds a root + intermediate so we can mint a credCert that chains to a
// root the verifier trusts — the same shape Apple uses, but with a synthetic root.
type testPKI struct {
	rootCert *x509.Certificate
	rootKey  *ecdsa.PrivateKey
	interDER []byte
	interX   *x509.Certificate
	interKey *ecdsa.PrivateKey
	roots    *x509.CertPool
}

func newTestPKI(t *testing.T) *testPKI {
	t.Helper()
	rootKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rootTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test App Attest Root"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	rootDER, _ := x509.CreateCertificate(rand.Reader, rootTmpl, rootTmpl, &rootKey.PublicKey, rootKey)
	rootCert, _ := x509.ParseCertificate(rootDER)

	interKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	interTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "Test App Attest CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	interDER, _ := x509.CreateCertificate(rand.Reader, interTmpl, rootCert, &interKey.PublicKey, rootKey)
	interX, _ := x509.ParseCertificate(interDER)

	pool := x509.NewCertPool()
	pool.AddCert(rootCert)
	return &testPKI{rootCert, rootKey, interDER, interX, interKey, pool}
}

func authDataFor(appID string, signCount uint32) []byte {
	sum := sha256.Sum256([]byte(appID))
	ad := make([]byte, 37)
	copy(ad[0:32], sum[:])
	ad[32] = 0x40
	ad[33] = byte(signCount >> 24)
	ad[34] = byte(signCount >> 16)
	ad[35] = byte(signCount >> 8)
	ad[36] = byte(signCount)
	return ad
}

func nonceExtensionValue(t *testing.T, nonce []byte) []byte {
	t.Helper()
	v, err := asn1.Marshal(struct {
		Nonce []byte `asn1:"tag:0,explicit"`
	}{Nonce: nonce})
	if err != nil {
		t.Fatalf("marshal nonce ext: %v", err)
	}
	return v
}

// mintCredCert mints the leaf credCert for `attestedKey`, embedding `nonce` in the
// Apple nonce extension, signed by the intermediate.
func (p *testPKI) mintCredCert(t *testing.T, attestedKey *ecdsa.PublicKey, nonce []byte) []byte {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "credCert"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		ExtraExtensions: []pkix.Extension{{
			Id:    nonceExtensionOID,
			Value: nonceExtensionValue(t, nonce),
		}},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, p.interX, attestedKey, p.interKey)
	if err != nil {
		t.Fatalf("create credCert: %v", err)
	}
	return der
}

// buildAttestation produces a full CBOR attestation object for the given key + challenge.
func (p *testPKI) buildAttestation(t *testing.T, attestedKey *ecdsa.PrivateKey, challenge []byte, signCount uint32) ([]byte, []byte) {
	t.Helper()
	ad := authDataFor(testAppID, signCount)
	clientHash := sha256.Sum256(challenge)
	nonce := sha256.Sum256(append(append([]byte{}, ad...), clientHash[:]...))
	credDER := p.mintCredCert(t, &attestedKey.PublicKey, nonce[:])

	obj := attestationObject{
		Fmt:      "apple-appattest",
		AttStmt:  attestationStmt{X5C: [][]byte{credDER, p.interDER}},
		AuthData: ad,
	}
	att, err := cbor.Marshal(obj)
	if err != nil {
		t.Fatalf("cbor attestation: %v", err)
	}
	pubX963 := elliptic963(&attestedKey.PublicKey)
	keyID := sha256.Sum256(pubX963)
	return att, keyID[:]
}

func TestVerifyAttestation_FullChainSucceeds(t *testing.T) {
	pki := newTestPKI(t)
	v := NewVerifier(testAppID, pki.roots)
	attestedKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	challenge := []byte("server-challenge-123")

	att, keyID := pki.buildAttestation(t, attestedKey, challenge, 0)
	res, err := v.VerifyAttestation(att, challenge, keyID)
	if err != nil {
		t.Fatalf("attestation should verify: %v", err)
	}
	if res.PublicKey.X.Cmp(attestedKey.PublicKey.X) != 0 {
		t.Fatal("returned key should equal the attested key")
	}
}

func TestVerifyAttestation_UntrustedRootFails(t *testing.T) {
	pki := newTestPKI(t)
	attestedKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	challenge := []byte("server-challenge-123")
	att, keyID := pki.buildAttestation(t, attestedKey, challenge, 0)

	// Verifier trusts an empty/foreign root pool → chain must fail.
	v := NewVerifier(testAppID, x509.NewCertPool())
	if _, err := v.VerifyAttestation(att, challenge, keyID); err == nil {
		t.Fatal("attestation with untrusted root must fail")
	}
}

func TestVerifyAttestation_WrongChallengeFailsNonce(t *testing.T) {
	pki := newTestPKI(t)
	v := NewVerifier(testAppID, pki.roots)
	attestedKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	att, keyID := pki.buildAttestation(t, attestedKey, []byte("real-challenge"), 0)

	if _, err := v.VerifyAttestation(att, []byte("attacker-challenge"), keyID); err == nil {
		t.Fatal("attestation must fail when the challenge (nonce) differs")
	}
}

func TestVerifyAttestation_WrongKeyIDFails(t *testing.T) {
	pki := newTestPKI(t)
	v := NewVerifier(testAppID, pki.roots)
	attestedKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	att, _ := pki.buildAttestation(t, attestedKey, []byte("c"), 0)

	bogus := sha256.Sum256([]byte("not the key"))
	if _, err := v.VerifyAttestation(att, []byte("c"), bogus[:]); err != ErrKeyIDMismatch {
		t.Fatalf("expected ErrKeyIDMismatch, got %v", err)
	}
}

// buildAssertion signs clientData with the attested key in the App Attest assertion shape.
func buildAssertion(t *testing.T, key *ecdsa.PrivateKey, clientData []byte, signCount uint32) []byte {
	t.Helper()
	ad := authDataFor(testAppID, signCount)
	clientHash := sha256.Sum256(clientData)
	nonce := sha256.Sum256(append(append([]byte{}, ad...), clientHash[:]...))
	sig, err := ecdsa.SignASN1(rand.Reader, key, nonce[:])
	if err != nil {
		t.Fatalf("sign assertion: %v", err)
	}
	b, _ := cbor.Marshal(struct {
		Signature []byte `cbor:"signature"`
		AuthData  []byte `cbor:"authenticatorData"`
	}{Signature: sig, AuthData: ad})
	return b
}

func TestVerifyAssertion_RoundTripAndSignCount(t *testing.T) {
	pki := newTestPKI(t)
	v := NewVerifier(testAppID, pki.roots)
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	clientData := []byte(`{"path":"/v1/foo","ts":123}`)

	// A valid assertion with an increasing counter is accepted.
	asrt := buildAssertion(t, key, clientData, 5)
	newCount, err := v.VerifyAssertion(asrt, clientData, &key.PublicKey, 4)
	if err != nil {
		t.Fatalf("assertion should verify: %v", err)
	}
	if newCount != 5 {
		t.Fatalf("expected new sign count 5, got %d", newCount)
	}

	// Replay/clone: a non-increasing counter is rejected.
	if _, err := v.VerifyAssertion(asrt, clientData, &key.PublicKey, 5); err != ErrSignCount {
		t.Fatalf("expected ErrSignCount on non-increasing counter, got %v", err)
	}

	// Tampered client data → signature fails.
	if _, err := v.VerifyAssertion(asrt, []byte("tampered"), &key.PublicKey, 4); err == nil {
		t.Fatal("assertion must fail when client data is tampered")
	}

	// Wrong key → signature fails.
	other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if _, err := v.VerifyAssertion(asrt, clientData, &other.PublicKey, 4); err == nil {
		t.Fatal("assertion must fail against a different key")
	}
}

func TestMemoryKeyStore(t *testing.T) {
	s := NewMemoryKeyStore()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err := s.Put("k1", KeyRecord{PublicKey: &key.PublicKey, SignCount: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateSignCount("k1", 7); err != nil {
		t.Fatal(err)
	}
	rec, ok, _ := s.Get("k1")
	if !ok || rec.SignCount != 7 {
		t.Fatalf("expected sign count 7, got %+v ok=%v", rec, ok)
	}
}
