package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	RecoverySessionTTL = 24 * time.Hour
	RecoveryShardThreshold = 2 // 2-of-3 Shamir threshold
	RecoveryMinContacts    = 3
)

// RecoveryMethod defines how account recovery is performed.
type RecoveryMethod string

const (
	RecoveryPhrase          RecoveryMethod = "recovery_phrase"
	RecoveryTrustedContacts RecoveryMethod = "trusted_contacts"
	RecoveryPhone           RecoveryMethod = "phone"
)

// RecoverySessionStatus tracks recovery progress.
type RecoverySessionStatus string

const (
	RecoveryPending   RecoverySessionStatus = "pending"
	RecoveryVerifying RecoverySessionStatus = "verifying"
	RecoveryComplete  RecoverySessionStatus = "complete"
	RecoveryExpired   RecoverySessionStatus = "expired"
	RecoveryFailed    RecoverySessionStatus = "failed"
)

// RecoverySession tracks an in-progress account recovery.
type RecoverySession struct {
	ID             string
	UserID         string
	Method         RecoveryMethod
	Status         RecoverySessionStatus
	RequiredSteps  []string
	CompletedSteps []string
	ShardsCollected int
	ShardsRequired  int
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

// TrustedContactShard represents a recovery shard from a trusted contact.
type TrustedContactShard struct {
	ContactUserID string
	SessionID     string
	Confirmed     bool
	ConfirmedAt   *time.Time
}

// RecoveryService manages account recovery flows.
type RecoveryService struct {
	mu       sync.RWMutex
	sessions map[string]*RecoverySession       // session_id -> session
	shards   map[string][]*TrustedContactShard // session_id -> shards
}

// NewRecoveryService creates a new recovery service.
func NewRecoveryService() *RecoveryService {
	return &RecoveryService{
		sessions: make(map[string]*RecoverySession),
		shards:   make(map[string][]*TrustedContactShard),
	}
}

// InitiateRecovery starts an account recovery session.
func (rs *RecoveryService) InitiateRecovery(userID string, method RecoveryMethod) (*RecoverySession, *AuthError) {
	sessionID := uuid.New().String()
	now := time.Now()

	var requiredSteps []string
	var shardsRequired int

	switch method {
	case RecoveryPhrase:
		requiredSteps = []string{"sign_challenge"}
		shardsRequired = 0
	case RecoveryTrustedContacts:
		requiredSteps = []string{"contact_confirmation", "register_passkey"}
		shardsRequired = RecoveryShardThreshold
	case RecoveryPhone:
		requiredSteps = []string{"verify_otp", "register_passkey"}
		shardsRequired = 0
	default:
		return nil, &AuthError{
			Code:       ErrCodeRecoveryInvalid,
			Message:    "Invalid recovery method",
			HTTPStatus: 400,
		}
	}

	session := &RecoverySession{
		ID:              sessionID,
		UserID:          userID,
		Method:          method,
		Status:          RecoveryPending,
		RequiredSteps:   requiredSteps,
		CompletedSteps:  []string{},
		ShardsCollected: 0,
		ShardsRequired:  shardsRequired,
		CreatedAt:       now,
		ExpiresAt:       now.Add(RecoverySessionTTL),
	}

	rs.mu.Lock()
	rs.sessions[sessionID] = session
	rs.mu.Unlock()

	return session, nil
}

// GetSession retrieves a recovery session.
func (rs *RecoveryService) GetSession(sessionID string) (*RecoverySession, *AuthError) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	session, exists := rs.sessions[sessionID]
	if !exists {
		return nil, NewAuthError(ErrCodeRecoveryInvalid, 400)
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = RecoveryExpired
		return nil, NewAuthError(ErrCodeRecoveryInvalid, 400)
	}

	return session, nil
}

// SubmitContactShard records a trusted contact's confirmation.
func (rs *RecoveryService) SubmitContactShard(sessionID, contactUserID string) (bool, *AuthError) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	session, exists := rs.sessions[sessionID]
	if !exists {
		return false, NewAuthError(ErrCodeRecoveryInvalid, 400)
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = RecoveryExpired
		return false, NewAuthError(ErrCodeRecoveryInvalid, 400)
	}

	if session.Method != RecoveryTrustedContacts {
		return false, &AuthError{
			Code:       ErrCodeRecoveryInvalid,
			Message:    "Session is not a trusted contacts recovery",
			HTTPStatus: 400,
		}
	}

	// Check for duplicate
	for _, shard := range rs.shards[sessionID] {
		if shard.ContactUserID == contactUserID {
			return false, nil // Already submitted
		}
	}

	now := time.Now()
	shard := &TrustedContactShard{
		ContactUserID: contactUserID,
		SessionID:     sessionID,
		Confirmed:     true,
		ConfirmedAt:   &now,
	}
	rs.shards[sessionID] = append(rs.shards[sessionID], shard)
	session.ShardsCollected++

	// Check if threshold is met
	if session.ShardsCollected >= session.ShardsRequired {
		session.Status = RecoveryVerifying
		session.CompletedSteps = append(session.CompletedSteps, "contact_confirmation")
		return true, nil // Threshold met
	}

	return false, nil
}

// CompleteStep marks a recovery step as complete.
func (rs *RecoveryService) CompleteStep(sessionID, step string) *AuthError {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	session, exists := rs.sessions[sessionID]
	if !exists {
		return NewAuthError(ErrCodeRecoveryInvalid, 400)
	}

	// Check if step is valid
	found := false
	for _, s := range session.RequiredSteps {
		if s == step {
			found = true
			break
		}
	}
	if !found {
		return &AuthError{
			Code:       ErrCodeRecoveryInvalid,
			Message:    "Invalid recovery step",
			HTTPStatus: 400,
		}
	}

	// Check if already completed
	for _, s := range session.CompletedSteps {
		if s == step {
			return nil // Already done
		}
	}

	session.CompletedSteps = append(session.CompletedSteps, step)

	// Check if all steps complete
	if len(session.CompletedSteps) == len(session.RequiredSteps) {
		session.Status = RecoveryComplete
	}

	return nil
}

// IsComplete checks if a recovery session has all steps completed.
func (rs *RecoveryService) IsComplete(sessionID string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	session, exists := rs.sessions[sessionID]
	if !exists {
		return false
	}
	return session.Status == RecoveryComplete
}

// SessionCount returns the number of active recovery sessions.
func (rs *RecoveryService) SessionCount() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.sessions)
}

// CleanExpired removes expired recovery sessions.
func (rs *RecoveryService) CleanExpired() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	now := time.Now()
	count := 0
	for id, session := range rs.sessions {
		if now.After(session.ExpiresAt) {
			delete(rs.sessions, id)
			delete(rs.shards, id)
			count++
		}
	}
	return count
}
