package crypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestX25519ChaChaSvc_RoundTrip(t *testing.T) {
	svc := NewX25519ChaChaSvc()

	recipPriv, recipPub, err := svc.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	plaintext := []byte("hello world — WO-13 X25519+ChaCha20Poly1305")
	msg, err := svc.EncryptWithKeyAgreement(plaintext, recipPub)
	if err != nil {
		t.Fatalf("EncryptWithKeyAgreement: %v", err)
	}

	if msg.Algorithm != X25519ChaChaAlgorithm {
		t.Errorf("algorithm: want %q, got %q", X25519ChaChaAlgorithm, msg.Algorithm)
	}

	got, err := svc.DecryptWithKeyAgreement(msg, recipPriv)
	if err != nil {
		t.Fatalf("DecryptWithKeyAgreement: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("plaintext mismatch: want %q, got %q", plaintext, got)
	}
}

func TestX25519ChaChaSvc_DifferentEphemeralEveryTime(t *testing.T) {
	svc := NewX25519ChaChaSvc()
	_, recipPub, _ := svc.GenerateKeyPair()

	msg1, _ := svc.EncryptWithKeyAgreement([]byte("same plaintext"), recipPub)
	msg2, _ := svc.EncryptWithKeyAgreement([]byte("same plaintext"), recipPub)

	if msg1.EphemeralPublicKey == msg2.EphemeralPublicKey {
		t.Error("ephemeral keys must differ across encryptions")
	}
	if msg1.Nonce == msg2.Nonce {
		t.Error("nonces must differ across encryptions")
	}
}

func TestX25519ChaChaSvc_TamperedCiphertextRejected(t *testing.T) {
	svc := NewX25519ChaChaSvc()
	recipPriv, recipPub, _ := svc.GenerateKeyPair()
	msg, _ := svc.EncryptWithKeyAgreement([]byte("tamper test"), recipPub)

	// Flip the first byte of the ciphertext
	ct, _ := base64.StdEncoding.DecodeString(msg.Ciphertext)
	ct[0] ^= 0xFF
	msg.Ciphertext = base64.StdEncoding.EncodeToString(ct)

	if _, err := svc.DecryptWithKeyAgreement(msg, recipPriv); err == nil {
		t.Error("tampered ciphertext should fail authentication")
	}
}

func TestX25519ChaChaSvc_WrongKeyRejected(t *testing.T) {
	svc := NewX25519ChaChaSvc()
	_, recipPub, _ := svc.GenerateKeyPair()
	wrongPriv, _, _ := svc.GenerateKeyPair()

	msg, _ := svc.EncryptWithKeyAgreement([]byte("wrong key"), recipPub)
	if _, err := svc.DecryptWithKeyAgreement(msg, wrongPriv); err == nil {
		t.Error("decryption with wrong key should fail")
	}
}

func TestX25519ChaChaSvc_InvalidPublicKeyRejected(t *testing.T) {
	svc := NewX25519ChaChaSvc()
	privKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Encrypt([]byte("test"), []byte("tooshort")); err == nil {
		t.Error("encrypt with short key should fail")
	}
	if _, err := svc.DeriveSharedKey(privKey, []byte("bad")); err == nil {
		t.Error("bad public key should fail")
	}
}

func TestX25519ChaChaSvc_EmptyPlaintext(t *testing.T) {
	svc := NewX25519ChaChaSvc()
	recipPriv, recipPub, _ := svc.GenerateKeyPair()
	msg, err := svc.EncryptWithKeyAgreement([]byte{}, recipPub)
	if err != nil {
		t.Fatalf("empty plaintext should be allowed: %v", err)
	}
	got, err := svc.DecryptWithKeyAgreement(msg, recipPriv)
	if err != nil {
		t.Fatalf("decrypt empty: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty plaintext, got %d bytes", len(got))
	}
}
