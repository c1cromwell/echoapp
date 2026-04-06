package crypto

import (
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- Test vector structures ---

type vectorFile struct {
	Vectors []testVector `json:"vectors"`
}

type testVector struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	SenderPrivateKey    string `json:"sender_private_key"`
	SenderPublicKey     string `json:"sender_public_key"`
	RecipientPrivateKey string `json:"recipient_private_key"`
	RecipientPublicKey  string `json:"recipient_public_key"`
	DerivedKey          string `json:"derived_key"`
	Plaintext           string `json:"plaintext"`
	Nonce               string `json:"nonce"`
	Ciphertext          string `json:"ciphertext"`
	Tag                 string `json:"tag"`
}

func loadVectors(t *testing.T) []testVector {
	t.Helper()

	// Find vectors.json relative to this test file
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	vectorPath := filepath.Join(projectRoot, "test", "crypto_vectors", "vectors.json")

	data, err := os.ReadFile(vectorPath)
	if err != nil {
		t.Fatalf("failed to read vectors.json: %v", err)
	}

	var vf vectorFile
	if err := json.Unmarshal(data, &vf); err != nil {
		t.Fatalf("failed to parse vectors.json: %v", err)
	}

	if len(vf.Vectors) == 0 {
		t.Fatal("no test vectors found in vectors.json")
	}
	return vf.Vectors
}

func mustDecode(t *testing.T, name, field, b64str string) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		t.Fatalf("[%s] failed to decode %s: %v", name, field, err)
	}
	return data
}

// --- Cross-platform validation tests ---

// TestVectors_KeyDerivation verifies that ECDH + HKDF produces
// the exact same derived key as recorded in the test vectors.
func TestVectors_KeyDerivation(t *testing.T) {
	ks := NewKinnamiService()
	vectors := loadVectors(t)

	for _, vec := range vectors {
		t.Run(vec.Name+"_key_derivation", func(t *testing.T) {
			senderPrivBytes := mustDecode(t, vec.Name, "sender_private_key", vec.SenderPrivateKey)
			recipientPubRaw := mustDecode(t, vec.Name, "recipient_public_key", vec.RecipientPublicKey)
			expectedKey := mustDecode(t, vec.Name, "derived_key", vec.DerivedKey)

			// Reconstruct sender's private key
			senderPriv, err := ecdh.P256().NewPrivateKey(senderPrivBytes)
			if err != nil {
				t.Fatalf("failed to parse sender private key: %v", err)
			}

			// Derive shared key
			derivedKey, err := ks.DeriveSharedKey(senderPriv, recipientPubRaw)
			if err != nil {
				t.Fatalf("DeriveSharedKey failed: %v", err)
			}

			if !bytesEqual(derivedKey, expectedKey) {
				t.Errorf("derived key mismatch\n  got:  %s\n  want: %s",
					base64.StdEncoding.EncodeToString(derivedKey),
					base64.StdEncoding.EncodeToString(expectedKey))
			}
		})
	}
}

// TestVectors_Decryption verifies that Go can decrypt ciphertext
// from the test vectors using the recorded key and nonce.
func TestVectors_Decryption(t *testing.T) {
	ks := NewKinnamiService()
	vectors := loadVectors(t)

	for _, vec := range vectors {
		t.Run(vec.Name+"_decrypt", func(t *testing.T) {
			derivedKey := mustDecode(t, vec.Name, "derived_key", vec.DerivedKey)

			msg := &KinnamiEncryptedMessage{
				Nonce:      vec.Nonce,
				Ciphertext: vec.Ciphertext,
				Tag:        vec.Tag,
				Algorithm:  KinnamiAlgorithm,
			}

			plaintext, err := ks.Decrypt(msg, derivedKey)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if string(plaintext) != vec.Plaintext {
				t.Errorf("plaintext mismatch\n  got:  %q\n  want: %q", string(plaintext), vec.Plaintext)
			}
		})
	}
}

// TestVectors_RecipientCanDecrypt verifies the full flow:
// recipient derives the same key from their private key + sender's public key,
// and can decrypt the message.
func TestVectors_RecipientCanDecrypt(t *testing.T) {
	ks := NewKinnamiService()
	vectors := loadVectors(t)

	for _, vec := range vectors {
		t.Run(vec.Name+"_recipient_decrypt", func(t *testing.T) {
			recipientPrivBytes := mustDecode(t, vec.Name, "recipient_private_key", vec.RecipientPrivateKey)
			senderPubRaw := mustDecode(t, vec.Name, "sender_public_key", vec.SenderPublicKey)

			// Recipient derives the same shared key from their private key + sender's public key
			recipientPriv, err := ecdh.P256().NewPrivateKey(recipientPrivBytes)
			if err != nil {
				t.Fatalf("failed to parse recipient private key: %v", err)
			}

			derivedKey, err := ks.DeriveSharedKey(recipientPriv, senderPubRaw)
			if err != nil {
				t.Fatalf("DeriveSharedKey failed: %v", err)
			}

			// Verify the derived key matches (ECDH is symmetric)
			expectedKey := mustDecode(t, vec.Name, "derived_key", vec.DerivedKey)
			if !bytesEqual(derivedKey, expectedKey) {
				t.Fatal("recipient derived a different key — ECDH symmetry broken")
			}

			// Decrypt with the derived key
			msg := &KinnamiEncryptedMessage{
				Nonce:      vec.Nonce,
				Ciphertext: vec.Ciphertext,
				Tag:        vec.Tag,
			}

			plaintext, err := ks.Decrypt(msg, derivedKey)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if string(plaintext) != vec.Plaintext {
				t.Errorf("plaintext mismatch\n  got:  %q\n  want: %q", string(plaintext), vec.Plaintext)
			}
		})
	}
}

// TestVectors_TamperedCiphertextFails verifies that modifying
// the ciphertext causes decryption to fail (GCM authentication).
func TestVectors_TamperedCiphertextFails(t *testing.T) {
	ks := NewKinnamiService()
	vectors := loadVectors(t)

	for _, vec := range vectors {
		t.Run(vec.Name+"_tampered", func(t *testing.T) {
			derivedKey := mustDecode(t, vec.Name, "derived_key", vec.DerivedKey)
			ctBytes := mustDecode(t, vec.Name, "ciphertext", vec.Ciphertext)

			// Skip empty ciphertext (empty plaintext → empty ciphertext)
			if len(ctBytes) == 0 {
				t.Skip("empty ciphertext, nothing to tamper")
			}

			// Flip a bit in the ciphertext
			ctBytes[0] ^= 0xFF
			tamperedCT := base64.StdEncoding.EncodeToString(ctBytes)

			msg := &KinnamiEncryptedMessage{
				Nonce:      vec.Nonce,
				Ciphertext: tamperedCT,
				Tag:        vec.Tag,
			}

			_, err := ks.Decrypt(msg, derivedKey)
			if err == nil {
				t.Error("expected decryption to fail with tampered ciphertext, but it succeeded")
			}
		})
	}
}

// TestKinnami_RoundTrip verifies that Go can encrypt and decrypt
// its own messages (not cross-platform, but ensures internal consistency).
func TestKinnami_RoundTrip(t *testing.T) {
	ks := NewKinnamiService()

	// Generate key pairs
	senderPriv, senderPub, err := ks.GenerateKeyPairP256()
	if err != nil {
		t.Fatalf("GenerateKeyPairP256 failed: %v", err)
	}

	recipientPriv, recipientPub, err := ks.GenerateKeyPairP256()
	if err != nil {
		t.Fatalf("GenerateKeyPairP256 failed: %v", err)
	}

	messages := []string{
		"Hello Echo!",
		"Unicode: 日本語 🎉",
		"",
		"A longer message with special chars: <>&\"'\\n\\t",
	}

	for _, plaintext := range messages {
		t.Run(plaintext, func(t *testing.T) {
			// Sender encrypts for recipient
			senderKey, err := ks.DeriveSharedKey(senderPriv, recipientPub)
			if err != nil {
				t.Fatalf("DeriveSharedKey failed: %v", err)
			}

			encrypted, err := ks.Encrypt([]byte(plaintext), senderKey)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Recipient derives the same key using sender's public key
			recipientKey, err := ks.DeriveSharedKey(recipientPriv, senderPub)
			if err != nil {
				t.Fatalf("DeriveSharedKey failed: %v", err)
			}

			// Keys should match (ECDH symmetry)
			if !bytesEqual(senderKey, recipientKey) {
				t.Fatal("ECDH keys don't match")
			}

			// Recipient decrypts
			decrypted, err := ks.Decrypt(encrypted, recipientKey)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if string(decrypted) != plaintext {
				t.Errorf("round-trip mismatch: got %q, want %q", string(decrypted), plaintext)
			}
		})
	}
}

// TestKinnami_EncryptWithKeyAgreement_RoundTrip tests the high-level
// EncryptWithKeyAgreement/DecryptWithKeyAgreement flow.
func TestKinnami_EncryptWithKeyAgreement_RoundTrip(t *testing.T) {
	ks := NewKinnamiService()

	_, recipientPub, err := ks.GenerateKeyPairP256()
	if err != nil {
		t.Fatal(err)
	}
	recipientPriv2, _, err := ks.GenerateKeyPairP256()
	if err != nil {
		t.Fatal(err)
	}

	// We need the actual recipient private key, so regenerate properly
	recipientPriv, recipientPubActual, err := ks.GenerateKeyPairP256()
	if err != nil {
		t.Fatal(err)
	}
	_ = recipientPub
	_ = recipientPriv2

	plaintext := "End-to-end encrypted message via Kinnami protocol"

	// Encrypt for recipient
	encrypted, err := ks.EncryptWithKeyAgreement([]byte(plaintext), recipientPubActual)
	if err != nil {
		t.Fatalf("EncryptWithKeyAgreement failed: %v", err)
	}

	if encrypted.Algorithm != KinnamiAlgorithmWithKA {
		t.Errorf("algorithm = %q, want %q", encrypted.Algorithm, KinnamiAlgorithmWithKA)
	}

	// Recipient decrypts
	decrypted, err := ks.DecryptWithKeyAgreement(encrypted, recipientPriv)
	if err != nil {
		t.Fatalf("DecryptWithKeyAgreement failed: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Errorf("got %q, want %q", string(decrypted), plaintext)
	}
}
