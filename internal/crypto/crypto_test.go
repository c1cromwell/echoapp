package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
)

func TestKeyDerivation(t *testing.T) {
	kds := NewKeyDerivationService()
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	derived, err := kds.DeriveApplicationKeys(masterKey, "test")
	if err != nil {
		t.Fatalf("derivation failed: %v", err)
	}
	if derived.SigningKey == nil || len(derived.EncryptionKey) != 32 {
		t.Error("invalid derived keys")
	}
}

func TestSigning(t *testing.T) {
	ss := NewSigningService()
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	msg := []byte("test")

	sig, err := ss.Sign(msg, privKey)
	if err != nil || len(sig) != 64 {
		t.Errorf("signing failed: %v, len=%d", err, len(sig))
	}

	if !ss.Verify(msg, sig, &privKey.PublicKey) {
		t.Error("verification failed")
	}
}

func TestHashing(t *testing.T) {
	hs := NewHashingService()
	data := []byte("test")
	salt := []byte("salt")

	hash := hs.HashWithSalt(data, salt)
	if len(hash) != 32 {
		t.Errorf("hash length=%d, want 32", len(hash))
	}

	hash2 := hs.HashWithSalt(data, salt)
	if !bytesEqual(hash, hash2) {
		t.Error("hash not deterministic")
	}
}

func TestCommitments(t *testing.T) {
	cs := NewCommitmentService()
	plain := []byte("secret")

	commit, err := cs.CreateMessageCommitment(plain)
	if err != nil || commit.Commitment == "" {
		t.Errorf("commitment failed: %v", err)
	}

	valid, _ := cs.VerifyCommitment(commit, plain)
	if !valid {
		t.Error("verification failed")
	}

	valid, _ = cs.VerifyCommitment(commit, []byte("wrong"))
	if valid {
		t.Error("should reject wrong plaintext")
	}
}
