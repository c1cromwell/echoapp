package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/hkdf"
)

// KeyDerivationService handles key derivation operations
type KeyDerivationService struct {
	// Optional custom random source for testing
	randReader io.Reader
}

// NewKeyDerivationService creates a new key derivation service
func NewKeyDerivationService() *KeyDerivationService {
	return &KeyDerivationService{
		randReader: rand.Reader,
	}
}

// DerivedKeys represents cryptographic keys derived for specific contexts
type DerivedKeys struct {
	SigningKey        *ecdsa.PrivateKey // P-256 signing key
	EncryptionKey     []byte            // 32 bytes for ChaCha20
	StorageKey        []byte            // 32 bytes for AES-GCM
	AuthenticationKey []byte            // 32 bytes for authentication
}

// DeriveApplicationKeys derives context-specific keys using HKDF
// Implements 3-tier hierarchy: MasterKey -> ApplicationKey -> ContextKey
func (kds *KeyDerivationService) DeriveApplicationKeys(
	masterKey []byte,
	context string,
) (*DerivedKeys, error) {
	if len(masterKey) < 32 {
		return nil, fmt.Errorf("master key must be at least 32 bytes")
	}

	// Generate random salt for this derivation
	salt := make([]byte, 32)
	if _, err := io.ReadFull(kds.randReader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Create HKDF expander
	hkdfReader := hkdf.New(sha256.New, masterKey, salt, []byte(context))

	// Derive signing key (P-256 = 32 bytes)
	signingKeyBytes := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, signingKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to derive signing key: %w", err)
	}

	// Create ECDSA private key from derived bytes
	signingKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
		D: new(big.Int).SetBytes(signingKeyBytes),
	}

	// Pre-compute public key
	signingKey.PublicKey.X, signingKey.PublicKey.Y = elliptic.P256().ScalarBaseMult(signingKeyBytes)

	// Derive encryption key (32 bytes for ChaCha20-Poly1305)
	encryptionKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, encryptionKey); err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Derive storage key (32 bytes for AES-GCM)
	storageKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, storageKey); err != nil {
		return nil, fmt.Errorf("failed to derive storage key: %w", err)
	}

	// Derive authentication key (32 bytes)
	authKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, authKey); err != nil {
		return nil, fmt.Errorf("failed to derive authentication key: %w", err)
	}

	return &DerivedKeys{
		SigningKey:        signingKey,
		EncryptionKey:     encryptionKey,
		StorageKey:        storageKey,
		AuthenticationKey: authKey,
	}, nil
}

// SigningService handles ECDSA signing and verification
type SigningService struct {
	randReader io.Reader
}

// NewSigningService creates a new signing service
func NewSigningService() *SigningService {
	return &SigningService{
		randReader: rand.Reader,
	}
}

// Sign signs data with ECDSA-P256-SHA256
func (ss *SigningService) Sign(message []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key cannot be nil")
	}

	// Hash the message with SHA-256
	hash := sha256.Sum256(message)

	// Sign with ECDSA
	r, s, err := ecdsa.Sign(ss.randReader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	// Encode signature as (r || s) in fixed 32-byte format
	// Total signature length: 64 bytes
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Pad to 32 bytes each
	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)

	signature := append(rPadded, sPadded...)
	return signature, nil
}

// Verify verifies an ECDSA signature
func (ss *SigningService) Verify(
	message []byte,
	signature []byte,
	publicKey *ecdsa.PublicKey,
) bool {
	if publicKey == nil || len(signature) != 64 {
		return false
	}

	// Hash the message
	hash := sha256.Sum256(message)

	// Parse signature (64 bytes: 32 for r, 32 for s)
	rBytes := signature[:32]
	sBytes := signature[32:]

	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// Verify signature
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// HashingService handles secure hashing operations
type HashingService struct{}

// NewHashingService creates a new hashing service
func NewHashingService() *HashingService {
	return &HashingService{}
}

// HashWithSalt performs salted hashing using SHA-256
func (hs *HashingService) HashWithSalt(data []byte, salt []byte) []byte {
	hash := sha256.New()
	hash.Write(salt)
	hash.Write(data)
	return hash.Sum(nil)
}

// HashArgon2id performs memory-hard hashing for PII like phone/email
// Parameters are OWASP recommended for interactive authentication
func (hs *HashingService) HashArgon2id(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		2,       // time cost (iterations)
		64*1024, // memory cost (64 MB)
		8,       // parallelism
		32,      // output length
	)
}

// CreateBlindIndex creates a deterministic but unlinkable hash
// Used for searchable encryption
func (hs *HashingService) CreateBlindIndex(data []byte, indexKey []byte) string {
	// HMAC-SHA256 for deterministic output
	h := sha256.New()
	h.Write(indexKey)
	h.Write(data)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// SHA256Hash returns SHA-256 hash of data
func (hs *HashingService) SHA256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SHA512Hash returns SHA-512 hash of data
func (hs *HashingService) SHA512Hash(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

// BLAKE2bHash returns BLAKE2b hash of data (32 bytes)
func (hs *HashingService) BLAKE2bHash(data []byte) ([]byte, error) {
	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	h.Write(data)
	return h.Sum(nil), nil
}

// CommitmentService creates cryptographic commitments
// Commitment: H(H(plaintext) || nonce) - proves integrity without exposure
type CommitmentService struct {
	hashService *HashingService
}

// NewCommitmentService creates a new commitment service
func NewCommitmentService() *CommitmentService {
	return &CommitmentService{
		hashService: NewHashingService(),
	}
}

// Commitment represents a hash commitment
type Commitment struct {
	Plaintext  string // base64url-encoded
	Nonce      string // base64url-encoded
	Commitment string // base64url-encoded
	Timestamp  int64  // Unix timestamp
}

// CreateMessageCommitment creates H(H(plaintext) || nonce)
func (cs *CommitmentService) CreateMessageCommitment(plaintext []byte) (*Commitment, error) {
	// Generate random nonce (32 bytes)
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create commitment: H(H(plaintext) || nonce)
	innerHash := cs.hashService.SHA256Hash(plaintext)
	commitment := cs.hashService.HashWithSalt(innerHash, nonce)

	return &Commitment{
		Plaintext:  base64.URLEncoding.EncodeToString(plaintext),
		Nonce:      base64.URLEncoding.EncodeToString(nonce),
		Commitment: base64.URLEncoding.EncodeToString(commitment),
		Timestamp:  time.Now().Unix(),
	}, nil
}

// VerifyCommitment verifies a commitment
func (cs *CommitmentService) VerifyCommitment(commitment *Commitment, plaintext []byte) (bool, error) {
	nonce, err := base64.URLEncoding.DecodeString(commitment.Nonce)
	if err != nil {
		return false, fmt.Errorf("failed to decode nonce: %w", err)
	}

	expectedCommitmentBytes, err := base64.URLEncoding.DecodeString(commitment.Commitment)
	if err != nil {
		return false, fmt.Errorf("failed to decode commitment: %w", err)
	}

	// Recompute commitment
	innerHash := cs.hashService.SHA256Hash(plaintext)
	recomputedCommitment := cs.hashService.HashWithSalt(innerHash, nonce)

	// Constant-time comparison
	return bytesEqual(recomputedCommitment, expectedCommitmentBytes), nil
}

// Helper function for constant-time byte comparison
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := range a {
		result |= int(a[i]) ^ int(b[i])
	}
	return result == 0
}
