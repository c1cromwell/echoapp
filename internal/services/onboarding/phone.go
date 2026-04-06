package onboarding

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"sync"
	"time"
)

const (
	OTPLength         = 6
	OTPExpiryDuration = 5 * time.Minute
	OTPCooldown       = 45 * time.Second
	MaxOTPAttempts    = 5
	MaxResends        = 3
)

// PhoneVerification tracks a phone verification attempt
type PhoneVerification struct {
	ID          string
	PhoneNumber string
	Code        string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Attempts    int
	Resends     int
	Verified    bool
	VerifiedAt  *time.Time
}

// PhoneVerificationService handles OTP generation and verification
type PhoneVerificationService struct {
	mu            sync.Mutex
	verifications map[string]*PhoneVerification // phone -> verification
}

// NewPhoneVerificationService creates a new phone verification service
func NewPhoneVerificationService() *PhoneVerificationService {
	return &PhoneVerificationService{
		verifications: make(map[string]*PhoneVerification),
	}
}

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// ValidatePhoneNumber checks if a phone number is in valid E.164 format
func ValidatePhoneNumber(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// SendCode generates and "sends" a verification code to the phone number
func (s *PhoneVerificationService) SendCode(phone string) (*PhoneVerification, error) {
	if !ValidatePhoneNumber(phone) {
		return nil, ErrInvalidPhoneNumber
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for existing verification and cooldown
	if existing, ok := s.verifications[phone]; ok {
		// If recently sent, enforce cooldown
		if time.Since(existing.CreatedAt) < OTPCooldown && !existing.Verified {
			return nil, ErrOTPAlreadySent
		}
		// Check resend limit
		if existing.Resends >= MaxResends {
			return nil, ErrOTPRateLimited
		}
	}

	code, err := generateOTP(OTPLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	now := time.Now()
	resends := 0
	if existing, ok := s.verifications[phone]; ok {
		resends = existing.Resends + 1
	}

	verification := &PhoneVerification{
		ID:          fmt.Sprintf("pv-%s-%d", phone[len(phone)-4:], now.UnixNano()),
		PhoneNumber: phone,
		Code:        code,
		CreatedAt:   now,
		ExpiresAt:   now.Add(OTPExpiryDuration),
		Resends:     resends,
	}

	s.verifications[phone] = verification
	return verification, nil
}

// VerifyCode validates the OTP code for a phone number
func (s *PhoneVerificationService) VerifyCode(phone, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	verification, ok := s.verifications[phone]
	if !ok {
		return ErrOTPInvalid
	}

	if verification.Verified {
		return nil // Already verified
	}

	if time.Now().After(verification.ExpiresAt) {
		return ErrOTPExpired
	}

	if verification.Attempts >= MaxOTPAttempts {
		return ErrOTPMaxAttempts
	}

	verification.Attempts++

	if verification.Code != code {
		return ErrOTPInvalid
	}

	now := time.Now()
	verification.Verified = true
	verification.VerifiedAt = &now
	return nil
}

// IsVerified checks if a phone number has been verified
func (s *PhoneVerificationService) IsVerified(phone string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.verifications[phone]
	return ok && v.Verified
}

// GetVerification returns the current verification state for a phone
func (s *PhoneVerificationService) GetVerification(phone string) (*PhoneVerification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.verifications[phone]
	if !ok {
		return nil, ErrOTPInvalid
	}
	return v, nil
}

// TimeUntilResend returns how long until a new code can be sent
func (s *PhoneVerificationService) TimeUntilResend(phone string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.verifications[phone]
	if !ok {
		return 0
	}

	elapsed := time.Since(v.CreatedAt)
	if elapsed >= OTPCooldown {
		return 0
	}
	return OTPCooldown - elapsed
}

// generateOTP creates a cryptographically random numeric code
func generateOTP(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
