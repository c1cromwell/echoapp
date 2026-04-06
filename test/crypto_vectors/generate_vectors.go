//go:build ignore

// This program generates cross-platform test vectors for the Kinnami
// encryption protocol. Run it once to create vectors.json, which is
// then used by both Go and Swift test suites.
//
// Usage: go run test/crypto_vectors/generate_vectors.go

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"
)

type TestVector struct {
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

type VectorFile struct {
	Protocol    string       `json:"protocol"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Parameters  interface{}  `json:"parameters"`
	Vectors     []TestVector `json:"vectors"`
}

func main() {
	testCases := []struct{ name, desc, plaintext string }{
		{"simple_hello", "Basic ASCII message", "Hello Echo!"},
		{"unicode_emoji", "Unicode with emoji", "Hey 👋 こんにちは"},
		{"empty_string", "Empty plaintext", ""},
		{"long_message", "Longer message simulating real chat", "This is a longer message that simulates a real conversation between two Echo users. It includes multiple sentences and should test that the encryption handles larger payloads correctly across both platforms."},
		{"json_payload", "JSON-structured content", `{"type":"text","content":"encrypted payload","timestamp":"2025-01-01T00:00:00Z"}`},
	}

	file := VectorFile{
		Protocol:    "Kinnami E2E Encryption",
		Version:     "1.0",
		Description: "Cross-platform test vectors for Go and Swift",
		Parameters: map[string]interface{}{
			"curve": "P-256", "kdf": "HKDF-SHA256",
			"salt": "ECHO-E2E-KINNAMI", "info": "message-encryption",
			"cipher": "AES-256-GCM", "key_bytes": 32, "nonce_bytes": 12, "tag_bytes": 16,
		},
	}

	for _, tc := range testCases {
		vec, err := generateVector(tc.name, tc.desc, tc.plaintext)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		file.Vectors = append(file.Vectors, *vec)
	}

	data, _ := json.MarshalIndent(file, "", "  ")
	os.WriteFile("test/crypto_vectors/vectors.json", data, 0644)
	fmt.Printf("Generated %d vectors -> test/crypto_vectors/vectors.json\n", len(file.Vectors))
}

func generateVector(name, description, plaintext string) (*TestVector, error) {
	senderPriv, _ := ecdh.P256().GenerateKey(rand.Reader)
	recipientPriv, _ := ecdh.P256().GenerateKey(rand.Reader)

	senderPubRaw := stripPrefix(senderPriv.PublicKey().Bytes())
	recipientPubRaw := stripPrefix(recipientPriv.PublicKey().Bytes())

	// ECDH + HKDF
	recipientPubKey, _ := ecdh.P256().NewPublicKey(append([]byte{0x04}, recipientPubRaw...))
	sharedSecret, _ := senderPriv.ECDH(recipientPubKey)

	hkdfR := hkdf.New(sha256.New, sharedSecret, []byte("ECHO-E2E-KINNAMI"), []byte("message-encryption"))
	derivedKey := make([]byte, 32)
	io.ReadFull(hkdfR, derivedKey)

	// AES-256-GCM
	block, _ := aes.NewCipher(derivedKey)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	tagSize := gcm.Overhead()
	ct := sealed[:len(sealed)-tagSize]
	tag := sealed[len(sealed)-tagSize:]

	return &TestVector{
		Name:                name,
		Description:         description,
		SenderPrivateKey:    b64(senderPriv.Bytes()),
		SenderPublicKey:     b64(senderPubRaw),
		RecipientPrivateKey: b64(recipientPriv.Bytes()),
		RecipientPublicKey:  b64(recipientPubRaw),
		DerivedKey:          b64(derivedKey),
		Plaintext:           plaintext,
		Nonce:               b64(nonce),
		Ciphertext:          b64(ct),
		Tag:                 b64(tag),
	}, nil
}

func stripPrefix(b []byte) []byte {
	if len(b) == 65 && b[0] == 0x04 {
		return b[1:]
	}
	return b
}

func b64(data []byte) string { return base64.StdEncoding.EncodeToString(data) }
