package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Ed25519KeyPair represents a public/private key pair
type Ed25519KeyPair struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() (*Ed25519KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Ed25519KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// Sign creates a signature for the given message
func (kp *Ed25519KeyPair) Sign(message []byte) ([]byte, error) {
	if kp.PrivateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	return ed25519.Sign(kp.PrivateKey, message), nil
}

// Verify verifies a signature for the given message
func (kp *Ed25519KeyPair) Verify(message, signature []byte) bool {
	if kp.PublicKey == nil {
		return false
	}
	return ed25519.Verify(kp.PublicKey, message, signature)
}

// PublicKeyHex returns the public key as a hex string
func (kp *Ed25519KeyPair) PublicKeyHex() string {
	return hex.EncodeToString(kp.PublicKey)
}

// PrivateKeyHex returns the private key as a hex string
func (kp *Ed25519KeyPair) PrivateKeyHex() string {
	return hex.EncodeToString(kp.PrivateKey)
}

// CryptoUtils provides cryptographic utilities
type CryptoUtils struct{}

// NewCryptoUtils creates a new CryptoUtils instance
func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{}
}

// GenerateKey generates a new Ed25519 key pair
func (c *CryptoUtils) GenerateKey() (*Ed25519KeyPair, error) {
	return GenerateKeyPair()
}

// SignMessage signs a message with the given private key
func (c *CryptoUtils) SignMessage(privateKey ed25519.PrivateKey, message []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	return ed25519.Sign(privateKey, message), nil
}

// VerifySignature verifies a signature with the given public key
func (c *CryptoUtils) VerifySignature(publicKey ed25519.PublicKey, message, signature []byte) bool {
	if publicKey == nil {
		return false
	}
	return ed25519.Verify(publicKey, message, signature)
}
