package crypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestPQHybrid_EncapDecapAgree(t *testing.T) {
	priv, bundle, err := GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	ct, ssA, err := HybridEncapsulate(bundle)
	if err != nil {
		t.Fatalf("encapsulate: %v", err)
	}
	ssB, err := HybridDecapsulate(priv, ct)
	if err != nil {
		t.Fatalf("decapsulate: %v", err)
	}
	if !bytes.Equal(ssA, ssB) {
		t.Fatalf("shared secrets differ:\n A=%x\n B=%x", ssA, ssB)
	}
	if len(ssA) != 32 {
		t.Fatalf("expected 32-byte secret, got %d", len(ssA))
	}
}

func TestPQHybrid_DistinctExchangesDiffer(t *testing.T) {
	_, bundle, _ := GenerateHybridKeyPair()
	_, ss1, _ := HybridEncapsulate(bundle)
	_, ss2, _ := HybridEncapsulate(bundle)
	if bytes.Equal(ss1, ss2) {
		t.Fatal("two encapsulations to the same bundle produced identical secrets")
	}
}

func TestPQHybrid_WrongKeyFails(t *testing.T) {
	_, bundle, _ := GenerateHybridKeyPair()
	otherPriv, _, _ := GenerateHybridKeyPair()
	ct, ssA, err := HybridEncapsulate(bundle)
	if err != nil {
		t.Fatalf("encapsulate: %v", err)
	}
	// Decapsulating with the wrong private key must not reproduce the secret.
	// (ML-KEM is designed to return a pseudo-random secret on mismatch, never error.)
	ssWrong, err := HybridDecapsulate(otherPriv, ct)
	if err != nil {
		t.Fatalf("decapsulate: %v", err)
	}
	if bytes.Equal(ssA, ssWrong) {
		t.Fatal("wrong key reproduced the shared secret")
	}
}

func TestPQHybrid_TamperedCiphertextDiverges(t *testing.T) {
	priv, bundle, _ := GenerateHybridKeyPair()
	ct, ssA, _ := HybridEncapsulate(bundle)

	raw, _ := base64.StdEncoding.DecodeString(ct.PQ)
	raw[0] ^= 0xff
	ct.PQ = base64.StdEncoding.EncodeToString(raw)

	ssB, err := HybridDecapsulate(priv, ct)
	if err != nil {
		t.Fatalf("decapsulate: %v", err)
	}
	if bytes.Equal(ssA, ssB) {
		t.Fatal("tampered PQ ciphertext still yielded the original secret")
	}
}

func TestPQHybrid_BundleRoundTrip(t *testing.T) {
	_, bundle, _ := GenerateHybridKeyPair()
	data, err := bundle.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	parsed, err := ParseHybridPublicBundle(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	ct, ssA, err := HybridEncapsulate(parsed)
	if err != nil {
		t.Fatalf("encapsulate parsed: %v", err)
	}
	if ct == nil || len(ssA) != 32 {
		t.Fatal("round-tripped bundle failed to encapsulate")
	}
}

// TestPQHybrid_BootstrapsRatchet proves the hybrid handshake yields a working
// Double Ratchet session end-to-end (PQ-protected session establishment).
func TestPQHybrid_BootstrapsRatchet(t *testing.T) {
	// Bob publishes a hybrid bundle + a ratchet pre-key.
	bobHybridPriv, bobBundle, err := GenerateHybridKeyPair()
	if err != nil {
		t.Fatalf("bob hybrid keygen: %v", err)
	}
	bobRatchet, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("bob ratchet keygen: %v", err)
	}
	bobRatchetPub := rawP256Pub(bobRatchet)

	// Alice initiates with PQ-hybrid handshake.
	alice, ct, err := NewRatchetInitiatorPQ(bobBundle, bobRatchetPub)
	if err != nil {
		t.Fatalf("alice init pq: %v", err)
	}
	// Bob completes the handshake by decapsulating Alice's ciphertext.
	bob, err := NewRatchetResponderPQ(bobHybridPriv, ct, bobRatchet)
	if err != nil {
		t.Fatalf("bob responder pq: %v", err)
	}

	// Exchange a couple of messages each way.
	m, _ := alice.Encrypt([]byte("pq hello"))
	if pt, err := bob.Decrypt(m); err != nil || string(pt) != "pq hello" {
		t.Fatalf("a->b: %q %v", pt, err)
	}
	r, _ := bob.Encrypt([]byte("pq reply"))
	if pt, err := alice.Decrypt(r); err != nil || string(pt) != "pq reply" {
		t.Fatalf("b->a: %q %v", pt, err)
	}
}
