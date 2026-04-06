// Package crypto provides the Go implementation of the Kinnami E2E encryption
// protocol, matching the Swift CryptoKit implementation in KinnamiEncryption.swift.
//
// Protocol:
//   1. Generate ephemeral P-256 key pair
//   2. ECDH key agreement with recipient's P-256 public key
//   3. HKDF-SHA256 key derivation (salt: "ECHO-E2E-KINNAMI", info: "message-encryption")
//   4. AES-256-GCM encryption with the derived 32-byte key
//
// The wire format uses base64-encoded fields to match iOS CryptoKit output.

package crypto

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

	"golang.org/x/crypto/hkdf"
)

const (
	// KinnamiSalt is the fixed HKDF salt matching the iOS implementation.
	KinnamiSalt = "ECHO-E2E-KINNAMI"
	// KinnamiInfo is the HKDF info string matching the iOS implementation.
	KinnamiInfo = "message-encryption"
	// KinnamiAlgorithm is the algorithm identifier for standard encryption.
	KinnamiAlgorithm = "AES-256-GCM"
	// KinnamiAlgorithmWithKA is the algorithm identifier for key-agreement encryption.
	KinnamiAlgorithmWithKA = "AES-256-GCM-KINNAMI"
)

// KinnamiEncryptedMessage represents an encrypted message (matches Swift EncryptedMessage).
type KinnamiEncryptedMessage struct {
	Nonce      string `json:"nonce"`      // base64-encoded 12-byte GCM nonce
	Ciphertext string `json:"ciphertext"` // base64-encoded ciphertext
	Tag        string `json:"tag"`        // base64-encoded 16-byte GCM tag
	Algorithm  string `json:"algorithm"`
}

// KinnamiEncryptedMessageWithKey includes the sender's ephemeral public key
// (matches Swift EncryptedMessageWithPublicKey).
type KinnamiEncryptedMessageWithKey struct {
	EphemeralPublicKey string `json:"ephemeralPublicKey"` // base64-encoded P-256 raw public key (65 bytes uncompressed)
	Nonce              string `json:"nonce"`
	Ciphertext         string `json:"ciphertext"`
	Tag                string `json:"tag"`
	Algorithm          string `json:"algorithm"`
}

// KinnamiService provides Kinnami E2E encryption compatible with the iOS implementation.
type KinnamiService struct{}

// NewKinnamiService creates a new Kinnami encryption service.
func NewKinnamiService() *KinnamiService {
	return &KinnamiService{}
}

// GenerateKeyPairP256 generates a new P-256 ECDH key pair.
// Returns (privateKey, publicKeyRawBytes).
// The public key raw bytes match CryptoKit's rawRepresentation (64 bytes: X || Y).
func (ks *KinnamiService) GenerateKeyPairP256() (*ecdh.PrivateKey, []byte, error) {
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate P-256 key: %w", err)
	}
	// PublicKey().Bytes() returns the uncompressed point (0x04 || X || Y = 65 bytes).
	// CryptoKit rawRepresentation is just X || Y (64 bytes).
	pubBytes := privateKey.PublicKey().Bytes()
	if len(pubBytes) == 65 && pubBytes[0] == 0x04 {
		pubBytes = pubBytes[1:] // strip the 0x04 prefix
	}
	return privateKey, pubBytes, nil
}

// DeriveSharedKey performs ECDH key agreement and HKDF derivation.
// This matches the Swift: sharedSecretFromKeyAgreement + hkdfDerivedSymmetricKey.
func (ks *KinnamiService) DeriveSharedKey(
	ourPrivateKey *ecdh.PrivateKey,
	theirPublicKeyRaw []byte,
) ([]byte, error) {
	// CryptoKit rawRepresentation is 64 bytes (X || Y).
	// Go ecdh expects uncompressed point: 0x04 || X || Y (65 bytes).
	var pubBytes []byte
	if len(theirPublicKeyRaw) == 64 {
		pubBytes = append([]byte{0x04}, theirPublicKeyRaw...)
	} else if len(theirPublicKeyRaw) == 65 && theirPublicKeyRaw[0] == 0x04 {
		pubBytes = theirPublicKeyRaw
	} else {
		return nil, fmt.Errorf("invalid public key length: %d (expected 64 or 65)", len(theirPublicKeyRaw))
	}

	theirPublicKey, err := ecdh.P256().NewPublicKey(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// ECDH
	sharedSecret, err := ourPrivateKey.ECDH(theirPublicKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	// HKDF-SHA256 derivation (matching CryptoKit parameters)
	hkdfReader := hkdf.New(
		sha256.New,
		sharedSecret,
		[]byte(KinnamiSalt),
		[]byte(KinnamiInfo),
	)

	derivedKey := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
		return nil, fmt.Errorf("HKDF derivation failed: %w", err)
	}

	return derivedKey, nil
}

// Encrypt encrypts plaintext with AES-256-GCM using the provided key.
func (ks *KinnamiService) Encrypt(plaintext []byte, key []byte) (*KinnamiEncryptedMessage, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// GCM Seal appends the tag to the ciphertext
	sealed := gcm.Seal(nil, nonce, plaintext, nil)

	// Split into ciphertext and tag (last 16 bytes)
	tagSize := gcm.Overhead() // 16
	ciphertext := sealed[:len(sealed)-tagSize]
	tag := sealed[len(sealed)-tagSize:]

	return &KinnamiEncryptedMessage{
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Tag:        base64.StdEncoding.EncodeToString(tag),
		Algorithm:  KinnamiAlgorithm,
	}, nil
}

// Decrypt decrypts an AES-256-GCM message using the provided key.
func (ks *KinnamiService) Decrypt(msg *KinnamiEncryptedMessage, key []byte) ([]byte, error) {
	nonce, err := base64.StdEncoding.DecodeString(msg.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(msg.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	tag, err := base64.StdEncoding.DecodeString(msg.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tag: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// GCM expects ciphertext || tag
	sealed := append(ciphertext, tag...)

	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (authentication error): %w", err)
	}

	return plaintext, nil
}

// EncryptWithKeyAgreement performs full Kinnami encryption:
// generates ephemeral key, does ECDH, derives key, encrypts.
func (ks *KinnamiService) EncryptWithKeyAgreement(
	plaintext []byte,
	recipientPublicKeyRaw []byte,
) (*KinnamiEncryptedMessageWithKey, error) {
	// Generate ephemeral key pair
	ephemeralPriv, ephemeralPubRaw, err := ks.GenerateKeyPairP256()
	if err != nil {
		return nil, err
	}

	// Derive shared key
	sharedKey, err := ks.DeriveSharedKey(ephemeralPriv, recipientPublicKeyRaw)
	if err != nil {
		return nil, err
	}

	// Encrypt
	encrypted, err := ks.Encrypt(plaintext, sharedKey)
	if err != nil {
		return nil, err
	}

	return &KinnamiEncryptedMessageWithKey{
		EphemeralPublicKey: base64.StdEncoding.EncodeToString(ephemeralPubRaw),
		Nonce:              encrypted.Nonce,
		Ciphertext:         encrypted.Ciphertext,
		Tag:                encrypted.Tag,
		Algorithm:          KinnamiAlgorithmWithKA,
	}, nil
}

// DecryptWithKeyAgreement decrypts using sender's ephemeral public key and our private key.
func (ks *KinnamiService) DecryptWithKeyAgreement(
	msg *KinnamiEncryptedMessageWithKey,
	ourPrivateKey *ecdh.PrivateKey,
) ([]byte, error) {
	ephemeralPubRaw, err := base64.StdEncoding.DecodeString(msg.EphemeralPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ephemeral public key: %w", err)
	}

	// Derive shared key
	sharedKey, err := ks.DeriveSharedKey(ourPrivateKey, ephemeralPubRaw)
	if err != nil {
		return nil, err
	}

	// Decrypt
	inner := &KinnamiEncryptedMessage{
		Nonce:      msg.Nonce,
		Ciphertext: msg.Ciphertext,
		Tag:        msg.Tag,
		Algorithm:  KinnamiAlgorithm,
	}
	return ks.Decrypt(inner, sharedKey)
}

// TestVectorToJSON serializes a test vector for cross-platform validation.
func TestVectorToJSON(vector interface{}) ([]byte, error) {
	return json.MarshalIndent(vector, "", "  ")
}
