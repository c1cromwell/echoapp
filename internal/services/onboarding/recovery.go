package onboarding

import (
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RecoveryMethodType represents the kind of recovery method
type RecoveryMethodType string

const (
	RecoveryEmail   RecoveryMethodType = "email"
	RecoveryWallet  RecoveryMethodType = "wallet"
	RecoveryContact RecoveryMethodType = "trusted_contact"
)

// RecoveryMethod represents a configured recovery option
type RecoveryMethod struct {
	ID         string
	UserID     string
	Type       RecoveryMethodType
	Value      string // email address, wallet address, or contact DID
	Verified   bool
	VerifiedAt *time.Time
	CreatedAt  time.Time
	Primary    bool
}

// RecoveryService manages account recovery methods
type RecoveryService struct {
	mu      sync.Mutex
	methods map[string]*RecoveryMethod // id -> method
	byUser  map[string][]string        // userID -> []methodID
}

// NewRecoveryService creates a new recovery service
func NewRecoveryService() *RecoveryService {
	return &RecoveryService{
		methods: make(map[string]*RecoveryMethod),
		byUser:  make(map[string][]string),
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// AddEmail sets up email-based recovery
func (s *RecoveryService) AddEmail(userID, email string) (*RecoveryMethod, error) {
	if userID == "" {
		return nil, ErrRecoveryEmailInvalid
	}
	if !ValidateEmail(email) {
		return nil, ErrRecoveryEmailInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	method := &RecoveryMethod{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      RecoveryEmail,
		Value:     email,
		CreatedAt: time.Now(),
		Primary:   !s.hasMethodsLocked(userID),
	}

	s.methods[method.ID] = method
	s.byUser[userID] = append(s.byUser[userID], method.ID)
	return method, nil
}

// AddWallet sets up wallet-based recovery
func (s *RecoveryService) AddWallet(userID, walletAddress string) (*RecoveryMethod, error) {
	if userID == "" || walletAddress == "" {
		return nil, ErrRecoveryWalletInvalid
	}
	if len(walletAddress) < 10 {
		return nil, ErrRecoveryWalletInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	method := &RecoveryMethod{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      RecoveryWallet,
		Value:     walletAddress,
		Verified:  true, // Wallet verification is implicit (signed tx)
		CreatedAt: time.Now(),
		Primary:   !s.hasMethodsLocked(userID),
	}

	now := time.Now()
	method.VerifiedAt = &now

	s.methods[method.ID] = method
	s.byUser[userID] = append(s.byUser[userID], method.ID)
	return method, nil
}

// AddTrustedContact sets up social recovery via a trusted contact
func (s *RecoveryService) AddTrustedContact(userID, contactDID string) (*RecoveryMethod, error) {
	if userID == "" || contactDID == "" {
		return nil, ErrRecoveryContactInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	method := &RecoveryMethod{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      RecoveryContact,
		Value:     contactDID,
		CreatedAt: time.Now(),
		Primary:   !s.hasMethodsLocked(userID),
	}

	s.methods[method.ID] = method
	s.byUser[userID] = append(s.byUser[userID], method.ID)
	return method, nil
}

// VerifyMethod marks a recovery method as verified
func (s *RecoveryService) VerifyMethod(methodID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	method, ok := s.methods[methodID]
	if !ok {
		return ErrRecoveryMethodInvalid
	}

	now := time.Now()
	method.Verified = true
	method.VerifiedAt = &now
	return nil
}

// GetUserMethods returns all recovery methods for a user
func (s *RecoveryService) GetUserMethods(userID string) []*RecoveryMethod {
	s.mu.Lock()
	defer s.mu.Unlock()

	methodIDs := s.byUser[userID]
	var methods []*RecoveryMethod
	for _, id := range methodIDs {
		if m, ok := s.methods[id]; ok {
			methods = append(methods, m)
		}
	}
	return methods
}

// GetMethod returns a specific recovery method
func (s *RecoveryService) GetMethod(methodID string) (*RecoveryMethod, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	method, ok := s.methods[methodID]
	if !ok {
		return nil, ErrRecoveryMethodInvalid
	}
	return method, nil
}

// RemoveMethod removes a recovery method
func (s *RecoveryService) RemoveMethod(methodID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	method, ok := s.methods[methodID]
	if !ok {
		return ErrRecoveryMethodInvalid
	}
	if method.UserID != userID {
		return ErrRecoveryMethodInvalid
	}

	delete(s.methods, methodID)

	// Remove from user's list
	ids := s.byUser[userID]
	for i, id := range ids {
		if id == methodID {
			s.byUser[userID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	return nil
}

// HasRecoveryMethod checks if a user has at least one recovery method
func (s *RecoveryService) HasRecoveryMethod(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.hasMethodsLocked(userID)
}

// HasVerifiedRecoveryMethod checks if a user has at least one verified recovery method
func (s *RecoveryService) HasVerifiedRecoveryMethod(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range s.byUser[userID] {
		if m, ok := s.methods[id]; ok && m.Verified {
			return true
		}
	}
	return false
}

func (s *RecoveryService) hasMethodsLocked(userID string) bool {
	return len(s.byUser[userID]) > 0
}
