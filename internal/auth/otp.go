package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	OTPLength        = 6
	OTPExpiry        = 10 * time.Minute
	OTPMaxAttempts   = 5
	OTPCooldown      = 60 * time.Second
	OTPBcryptCost    = 10
	MaxOTPPerPhoneHr = 5
)

// OTPSession tracks an in-progress OTP verification.
type OTPSession struct {
	VerificationID string
	PhoneHash      string
	CodeHash       string // bcrypt hash of the OTP code
	Attempts       int
	CreatedAt      time.Time
	ExpiresAt      time.Time
	Verified       bool
}

// OTPService handles OTP generation, hashing, and verification.
type OTPService struct {
	mu       sync.RWMutex
	sessions map[string]*OTPSession // key: verification_id
	phoneLimits map[string][]time.Time // key: phone_hash -> timestamps of OTP sends
}

// NewOTPService creates a new OTP service.
func NewOTPService() *OTPService {
	return &OTPService{
		sessions:    make(map[string]*OTPSession),
		phoneLimits: make(map[string][]time.Time),
	}
}

// GenerateOTP creates a cryptographically secure 6-digit code.
func GenerateOTP() (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(OTPLength)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate OTP: %w", err)
	}
	return fmt.Sprintf("%0*d", OTPLength, n.Int64()), nil
}

// HashOTP creates a bcrypt hash of the OTP code.
func HashOTP(code string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(code), OTPBcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash OTP: %w", err)
	}
	return string(hash), nil
}

// VerifyOTPHash checks a code against a bcrypt hash.
func VerifyOTPHash(code, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil
}

// CheckPhoneRateLimit returns true if the phone has exceeded OTP send limits.
func (s *OTPService) CheckPhoneRateLimit(phoneHash string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)

	// Remove expired entries
	timestamps := s.phoneLimits[phoneHash]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	s.phoneLimits[phoneHash] = valid

	return len(valid) >= MaxOTPPerPhoneHr
}

// RecordOTPSend records that an OTP was sent to this phone.
func (s *OTPService) RecordOTPSend(phoneHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phoneLimits[phoneHash] = append(s.phoneLimits[phoneHash], time.Now())
}

// CreateSession stores a new OTP verification session.
func (s *OTPService) CreateSession(verificationID, phoneHash, codeHash string) *OTPSession {
	session := &OTPSession{
		VerificationID: verificationID,
		PhoneHash:      phoneHash,
		CodeHash:       codeHash,
		Attempts:       0,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(OTPExpiry),
		Verified:       false,
	}

	s.mu.Lock()
	s.sessions[verificationID] = session
	s.mu.Unlock()

	return session
}

// GetSession retrieves an OTP session by verification ID.
func (s *OTPService) GetSession(verificationID string) *OTPSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[verificationID]
}

// VerifyCode checks the OTP code against the stored session.
// Returns (success, error). On success the session is marked verified.
func (s *OTPService) VerifyCode(verificationID, code string) (bool, *AuthError) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[verificationID]
	if !exists {
		return false, NewAuthError(ErrCodeInvalidOTP, 400)
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, verificationID)
		return false, NewAuthError(ErrCodeInvalidOTP, 400)
	}

	// Check attempts
	session.Attempts++
	if session.Attempts > OTPMaxAttempts {
		delete(s.sessions, verificationID)
		return false, NewAuthError(ErrCodeInvalidOTP, 400)
	}

	// Verify code
	if !VerifyOTPHash(code, session.CodeHash) {
		return false, NewAuthError(ErrCodeInvalidOTP, 400)
	}

	session.Verified = true
	return true, nil
}

// InvalidateSession removes an OTP session.
func (s *OTPService) InvalidateSession(verificationID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, verificationID)
}

// CleanExpired removes all expired sessions.
func (s *OTPService) CleanExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
			count++
		}
	}
	return count
}

// SessionCount returns the number of active sessions.
func (s *OTPService) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
