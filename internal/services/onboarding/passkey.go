package onboarding

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PasskeyType represents the biometric type used
type PasskeyType string

const (
	PasskeyFaceID      PasskeyType = "face_id"
	PasskeyTouchID     PasskeyType = "touch_id"
	PasskeyFingerprint PasskeyType = "fingerprint"
	PasskeyPIN         PasskeyType = "pin"
)

// PasskeyCredential represents a registered passkey
type PasskeyCredential struct {
	ID              string
	UserID          string
	CredentialID    string // WebAuthn credential ID
	PublicKey       []byte
	PasskeyType     PasskeyType
	DeviceInfo      string
	CreatedAt       time.Time
	LastUsedAt      *time.Time
	AttestationData []byte
}

// PasskeyChallenge represents a WebAuthn registration challenge
type PasskeyChallenge struct {
	ID        string
	UserID    string
	Challenge []byte
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

// PasskeyService manages passkey registration and verification
type PasskeyService struct {
	mu          sync.Mutex
	credentials map[string]*PasskeyCredential // credentialID -> credential
	challenges  map[string]*PasskeyChallenge  // challengeID -> challenge
	userKeys    map[string][]string           // userID -> []credentialID
}

// NewPasskeyService creates a new passkey service
func NewPasskeyService() *PasskeyService {
	return &PasskeyService{
		credentials: make(map[string]*PasskeyCredential),
		challenges:  make(map[string]*PasskeyChallenge),
		userKeys:    make(map[string][]string),
	}
}

// CreateChallenge generates a new WebAuthn registration challenge
func (s *PasskeyService) CreateChallenge(userID string) (*PasskeyChallenge, error) {
	if userID == "" {
		return nil, ErrPasskeyInvalidData
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	challengeBytes := make([]byte, 32)
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())))
	copy(challengeBytes, hash[:])

	challenge := &PasskeyChallenge{
		ID:        uuid.New().String(),
		UserID:    userID,
		Challenge: challengeBytes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	s.challenges[challenge.ID] = challenge
	return challenge, nil
}

// RegisterCredential registers a new passkey credential after successful biometric auth
func (s *PasskeyService) RegisterCredential(challengeID string, credentialID string, publicKey []byte, passkeyType PasskeyType, deviceInfo string) (*PasskeyCredential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[challengeID]
	if !ok {
		return nil, ErrPasskeyInvalidData
	}

	if challenge.Used {
		return nil, ErrPasskeyAlreadyExists
	}

	if time.Now().After(challenge.ExpiresAt) {
		return nil, ErrPasskeyCreationFailed
	}

	if len(publicKey) == 0 || credentialID == "" {
		return nil, ErrPasskeyInvalidData
	}

	// Check if user already has a passkey
	if keys, exists := s.userKeys[challenge.UserID]; exists && len(keys) > 0 {
		return nil, ErrPasskeyAlreadyExists
	}

	credential := &PasskeyCredential{
		ID:           uuid.New().String(),
		UserID:       challenge.UserID,
		CredentialID: credentialID,
		PublicKey:    publicKey,
		PasskeyType:  passkeyType,
		DeviceInfo:   deviceInfo,
		CreatedAt:    time.Now(),
	}

	challenge.Used = true
	s.credentials[credentialID] = credential
	s.userKeys[challenge.UserID] = append(s.userKeys[challenge.UserID], credentialID)

	return credential, nil
}

// GetCredential retrieves a passkey credential by its credential ID
func (s *PasskeyService) GetCredential(credentialID string) (*PasskeyCredential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cred, ok := s.credentials[credentialID]
	if !ok {
		return nil, ErrPasskeyInvalidData
	}
	return cred, nil
}

// HasPasskey checks if a user has a registered passkey
func (s *PasskeyService) HasPasskey(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys, exists := s.userKeys[userID]
	return exists && len(keys) > 0
}

// GetUserCredentials returns all passkey credentials for a user
func (s *PasskeyService) GetUserCredentials(userID string) []*PasskeyCredential {
	s.mu.Lock()
	defer s.mu.Unlock()

	credIDs := s.userKeys[userID]
	var creds []*PasskeyCredential
	for _, id := range credIDs {
		if cred, ok := s.credentials[id]; ok {
			creds = append(creds, cred)
		}
	}
	return creds
}

// GenerateDID generates a DID from the passkey's public key
func GenerateDID(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return "did:echo:" + hex.EncodeToString(hash[:16])
}
