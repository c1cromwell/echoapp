package credentials

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// CryptoUtils provides cryptographic operations for credentials
type CryptoUtils struct {
	randSource io.Reader
}

// NewCryptoUtils creates new crypto utilities
func NewCryptoUtils() *CryptoUtils {
	return &CryptoUtils{
		randSource: rand.Reader,
	}
}

// SignMessage signs a message with Ed25519
func (c *CryptoUtils) SignMessage(privateKeyBase64 string, message []byte) (string, error) {
	// Decode base64 private key
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to decode private key",
			err.Error(),
		)
	}

	// Validate private key length (Ed25519 private keys are 64 bytes)
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return "", NewCredentialError(
			ErrCodeInvalidCredential,
			"invalid private key size for Ed25519",
		)
	}

	// Create Ed25519 private key
	privateKey := ed25519.PrivateKey(privateKeyBytes)

	// Sign the message
	signature := ed25519.Sign(privateKey, message)

	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies a signature with Ed25519
func (c *CryptoUtils) VerifySignature(publicKeyBase64 string, message []byte, signatureBase64 string) (bool, error) {
	// Decode base64 public key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return false, NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to decode public key",
			err.Error(),
		)
	}

	// Decode base64 signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, NewCredentialErrorWithDetails(
			ErrCodeInvalidProof,
			"failed to decode signature",
			err.Error(),
		)
	}

	// Create Ed25519 public key
	publicKey := ed25519.PublicKey(publicKeyBytes)

	// Verify signature
	valid := ed25519.Verify(publicKey, message, signatureBytes)
	return valid, nil
}

// GenerateKeyPair generates Ed25519 key pair
func (c *CryptoUtils) GenerateKeyPair() (publicKeyBase64, privateKeyBase64 string, err error) {
	publicKey, privateKey, err := ed25519.GenerateKey(c.randSource)
	if err != nil {
		return "", "", NewCredentialErrorWithDetails(
			ErrCodeIssuanceFailed,
			"failed to generate key pair",
			err.Error(),
		)
	}

	publicKeyBase64 = base64.StdEncoding.EncodeToString(publicKey)
	privateKeyBytes := make([]byte, len(privateKey))
	copy(privateKeyBytes, privateKey)
	privateKeyBase64 = base64.StdEncoding.EncodeToString(privateKeyBytes)

	return publicKeyBase64, privateKeyBase64, nil
}

// GenerateChallenge generates a random challenge
func (c *CryptoUtils) GenerateChallenge(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	challenge := make([]byte, length)
	_, err := c.randSource.Read(challenge)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeIssuanceFailed,
			"failed to generate challenge",
			err.Error(),
		)
	}

	return base64.RawURLEncoding.EncodeToString(challenge), nil
}

// GenerateNonce generates a random nonce
func (c *CryptoUtils) GenerateNonce(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	nonce := make([]byte, length)
	_, err := c.randSource.Read(nonce)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeIssuanceFailed,
			"failed to generate nonce",
			err.Error(),
		)
	}

	return hex.EncodeToString(nonce), nil
}

// HashMessage creates SHA256 hash of message
func (c *CryptoUtils) HashMessage(message []byte) string {
	hash := sha256.Sum256(message)
	return hex.EncodeToString(hash[:])
}

// HashMessageSHA512 creates SHA512 hash of message
func (c *CryptoUtils) HashMessageSHA512(message []byte) string {
	hash := sha512.Sum512(message)
	return hex.EncodeToString(hash[:])
}

// ExtractPublicKeyFromPrivateKey extracts public key from private key
func (c *CryptoUtils) ExtractPublicKeyFromPrivateKey(privateKeyBase64 string) (string, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to decode private key",
			err.Error(),
		)
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return base64.StdEncoding.EncodeToString(publicKey), nil
}

// IsValidEd25519PublicKey validates Ed25519 public key format
func (c *CryptoUtils) IsValidEd25519PublicKey(publicKeyBase64 string) bool {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return false
	}
	return len(publicKeyBytes) == ed25519.PublicKeySize
}

// IsValidEd25519PrivateKey validates Ed25519 private key format
func (c *CryptoUtils) IsValidEd25519PrivateKey(privateKeyBase64 string) bool {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return false
	}
	return len(privateKeyBytes) == ed25519.PrivateKeySize
}

// CreateJWSSignature creates a JWS signature (JWT-compatible)
func (c *CryptoUtils) CreateJWSSignature(header, payload, privateKeyBase64 string) (string, error) {
	// Encode header and payload as base64url
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(payload))

	// Create signature input
	signatureInput := fmt.Sprintf("%s.%s", headerB64, payloadB64)

	// Sign
	signature, err := c.SignMessage(privateKeyBase64, []byte(signatureInput))
	if err != nil {
		return "", err
	}

	// Decode signature from standard base64 to raw base64url
	sigBytes, _ := base64.StdEncoding.DecodeString(signature)
	sigB64url := base64.RawURLEncoding.EncodeToString(sigBytes)

	// Return complete JWS
	return fmt.Sprintf("%s.%s", signatureInput, sigB64url), nil
}

// VerifyJWSSignature verifies a JWS signature
func (c *CryptoUtils) VerifyJWSSignature(jws, publicKeyBase64 string) (bool, error) {
	// Split JWS into parts
	parts := make([]string, 3)
	j := 0
	for _, ch := range jws {
		if ch == '.' {
			j++
		} else if j < 3 {
			parts[j] += string(ch)
		}
	}

	if len(parts) != 3 {
		return false, NewCredentialError(ErrCodeInvalidProof, "invalid JWS format")
	}

	// Decode signature from base64url to standard base64
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false, NewCredentialErrorWithDetails(ErrCodeInvalidProof, "failed to decode signature", err.Error())
	}
	sigB64 := base64.StdEncoding.EncodeToString(sigBytes)

	// Verify
	return c.VerifySignature(publicKeyBase64, []byte(fmt.Sprintf("%s.%s", parts[0], parts[1])), sigB64)
}

// SignCredentialProof signs a credential proof
func (c *CryptoUtils) SignCredentialProof(credentialJSON []byte, privateKeyBase64 string) (proofValue string, err error) {
	// Create hash of credential
	credentialHash := c.HashMessageSHA512(credentialJSON)

	// Sign the hash
	hashBytes := []byte(credentialHash)
	signature, err := c.SignMessage(privateKeyBase64, hashBytes)
	if err != nil {
		return "", err
	}

	return signature, nil
}

// VerifyCredentialProof verifies a credential proof
func (c *CryptoUtils) VerifyCredentialProof(credentialJSON []byte, proofValue, publicKeyBase64 string) (bool, error) {
	// Create hash of credential
	credentialHash := c.HashMessageSHA512(credentialJSON)

	// Verify signature
	return c.VerifySignature(publicKeyBase64, []byte(credentialHash), proofValue)
}

// ECDSAUtils provides ECDSA operations (for future use)
type ECDSAUtils struct {
	randSource io.Reader
}

// NewECDSAUtils creates new ECDSA utilities
func NewECDSAUtils() *ECDSAUtils {
	return &ECDSAUtils{
		randSource: rand.Reader,
	}
}

// SignECDSA signs message with ECDSA (placeholder for future use)
func (e *ECDSAUtils) SignECDSA(privateKey *ecdsa.PrivateKey, message []byte) (string, error) {
	// This is a placeholder for ECDSA support
	// Full implementation would follow similar pattern to Ed25519
	return "", fmt.Errorf("ECDSA signing not yet implemented")
}

// VerifyECDSA verifies ECDSA signature (placeholder for future use)
func (e *ECDSAUtils) VerifyECDSA(publicKey *ecdsa.PublicKey, message []byte, signature string) (bool, error) {
	// This is a placeholder for ECDSA support
	return false, fmt.Errorf("ECDSA verification not yet implemented")
}

// GetHashAlgorithm returns hash algorithm for signing method
func GetHashAlgorithm(signingMethod string) crypto.Hash {
	switch signingMethod {
	case "Ed25519":
		return crypto.SHA512
	case "ECDSA":
		return crypto.SHA256
	default:
		return crypto.SHA256
	}
}

// GetSignatureAlgorithm returns signature algorithm name
func GetSignatureAlgorithm(signingMethod string) string {
	switch signingMethod {
	case "Ed25519":
		return "EdDSA"
	case "ECDSA":
		return "ES256"
	default:
		return "EdDSA"
	}
}
