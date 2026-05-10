package crypto

// X25519 ECDH + ChaCha20-Poly1305 message encryption (WO-13 canonical spec).
//
// This is the wire-protocol encryption for iOS ↔ backend message payloads:
//   1. iOS generates ephemeral X25519 key pair
//   2. iOS ECDH with backend's static X25519 public key
//   3. HKDF-SHA256 key derivation (salt: "ECHO-X25519-KINNAMI", info: "message-encryption-x25519")
//   4. ChaCha20-Poly1305 encryption of the message blob
//
// The backend operates as a content-blind relay — it never decrypts message
// blobs between users; X25519/ChaCha20 is used only for iOS→backend service
// RPC encryption (key exchange, status APIs, etc.).

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

const (
	X25519ChaChaSalt      = "ECHO-X25519-KINNAMI"
	X25519ChaChaInfo      = "message-encryption-x25519"
	X25519ChaChaAlgorithm = "X25519-ChaCha20Poly1305"
)

// X25519ChaChaMessage is the wire format for an X25519+ChaCha20Poly1305-encrypted blob.
type X25519ChaChaMessage struct {
	EphemeralPublicKey string `json:"ephemeral_public_key"` // base64, 32-byte X25519 raw pubkey
	Nonce              string `json:"nonce"`                // base64, 12-byte ChaCha20 nonce
	Ciphertext         string `json:"ciphertext"`           // base64, ciphertext+poly1305 tag
	Algorithm          string `json:"algorithm"`            // always X25519ChaChaAlgorithm
}

// X25519ChaChaSvc implements WO-13 X25519+ChaCha20-Poly1305 encryption.
type X25519ChaChaSvc struct{}

// NewX25519ChaChaSvc creates a new service instance.
func NewX25519ChaChaSvc() *X25519ChaChaSvc { return &X25519ChaChaSvc{} }

// GenerateKeyPair generates a fresh X25519 key pair.
// Returns (privateKey, 32-byte raw public key bytes).
func (s *X25519ChaChaSvc) GenerateKeyPair() (*ecdh.PrivateKey, []byte, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("x25519 key gen: %w", err)
	}
	return priv, priv.PublicKey().Bytes(), nil
}

// DeriveSharedKey performs X25519 ECDH + HKDF-SHA256 derivation to produce a
// 32-byte ChaCha20-Poly1305 symmetric key.
func (s *X25519ChaChaSvc) DeriveSharedKey(ourPriv *ecdh.PrivateKey, theirPubBytes []byte) ([]byte, error) {
	theirPub, err := ecdh.X25519().NewPublicKey(theirPubBytes)
	if err != nil {
		return nil, fmt.Errorf("parse X25519 public key: %w", err)
	}
	shared, err := ourPriv.ECDH(theirPub)
	if err != nil {
		return nil, fmt.Errorf("X25519 ECDH: %w", err)
	}
	r := hkdf.New(sha256.New, shared, []byte(X25519ChaChaSalt), []byte(X25519ChaChaInfo))
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("HKDF derivation: %w", err)
	}
	return key, nil
}

// Encrypt encrypts plaintext with ChaCha20-Poly1305 using the provided 32-byte key.
// Returns (nonce, ciphertext+tag).
func (s *X25519ChaChaSvc) Encrypt(plaintext, key []byte) (nonce, ciphertext []byte, err error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, fmt.Errorf("chacha20poly1305 init: %w", err)
	}
	nonce = make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("nonce gen: %w", err)
	}
	ciphertext = aead.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt decrypts a ChaCha20-Poly1305 ciphertext (includes Poly1305 tag at the end).
func (s *X25519ChaChaSvc) Decrypt(nonce, ciphertext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("chacha20poly1305 init: %w", err)
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("chacha20poly1305 decrypt: %w", err)
	}
	return plaintext, nil
}

// EncryptWithKeyAgreement is the full one-shot iOS→backend encryption:
// generates an ephemeral X25519 key, performs ECDH with the recipient's
// static public key, derives a ChaCha20-Poly1305 key, and encrypts.
func (s *X25519ChaChaSvc) EncryptWithKeyAgreement(plaintext, recipientPubBytes []byte) (*X25519ChaChaMessage, error) {
	ephPriv, ephPub, err := s.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	key, err := s.DeriveSharedKey(ephPriv, recipientPubBytes)
	if err != nil {
		return nil, err
	}
	nonce, ct, err := s.Encrypt(plaintext, key)
	if err != nil {
		return nil, err
	}
	return &X25519ChaChaMessage{
		EphemeralPublicKey: base64.StdEncoding.EncodeToString(ephPub),
		Nonce:              base64.StdEncoding.EncodeToString(nonce),
		Ciphertext:         base64.StdEncoding.EncodeToString(ct),
		Algorithm:          X25519ChaChaAlgorithm,
	}, nil
}

// DecryptWithKeyAgreement decrypts a message using the recipient's static X25519 private key.
func (s *X25519ChaChaSvc) DecryptWithKeyAgreement(msg *X25519ChaChaMessage, ourPriv *ecdh.PrivateKey) ([]byte, error) {
	ephPubBytes, err := base64.StdEncoding.DecodeString(msg.EphemeralPublicKey)
	if err != nil {
		return nil, fmt.Errorf("decode ephemeral pubkey: %w", err)
	}
	key, err := s.DeriveSharedKey(ourPriv, ephPubBytes)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(msg.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}
	ct, err := base64.StdEncoding.DecodeString(msg.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	return s.Decrypt(nonce, ct, key)
}
