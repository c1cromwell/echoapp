package recovery

import (
	"errors"
	"fmt"
	"time"
)

const (
	DefaultThreshold = 2
	DefaultTotal     = 3

	RoleDevice  = "device"
	RoleContact = "contact"
	RoleOrg     = "org"

	StatusPending  = "pending"
	StatusActive   = "active"
	StatusRevoked  = "revoked"

	SessionInitiated  = "initiated"
	SessionCompleted  = "completed"
	SessionExpired    = "expired"
	SessionCancelled  = "cancelled"
)

var (
	ErrNotConfigured     = errors.New("recovery not configured for holder")
	ErrInvalidPolicy     = errors.New("invalid recovery policy")
	ErrInvalidShareholder = errors.New("invalid shareholder metadata")
	ErrSessionNotFound   = errors.New("recovery session not found")
	ErrSessionExpired    = errors.New("recovery session expired")
	ErrForbiddenField    = errors.New("server cannot accept share material or recovery secrets")
)

// Policy is the holder's M-of-N recovery configuration (metadata only).
type Policy struct {
	HolderDID string    `json:"holder_did"`
	Threshold int       `json:"threshold_m"`
	Total     int       `json:"total_n"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Shareholder is share metadata — never Shamir share bytes (WO-296 / ADR 0004).
type Shareholder struct {
	ShareID      string    `json:"share_id"`
	HolderDID    string    `json:"holder_did"`
	ShareIndex   int       `json:"share_index"`
	GuardianDID  string    `json:"guardian_did"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	GuardianVCID string    `json:"guardian_vc_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session tracks an in-progress social recovery attempt.
type Session struct {
	SessionID         string     `json:"session_id"`
	HolderDID         string     `json:"holder_did"`
	Status            string     `json:"status"`
	RequiredShares    int        `json:"required_shares"`
	ExpiresAt         time.Time  `json:"expires_at"`
	RootKeyCommitment string     `json:"root_key_commitment,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// SetupRequest registers recovery policy + shareholder metadata after client-side Shamir split.
type SetupRequest struct {
	Threshold    int                 `json:"threshold_m"`
	Total        int                 `json:"total_n"`
	Shareholders []ShareholderInput  `json:"shareholders"`
}

type ShareholderInput struct {
	ShareIndex   int    `json:"share_index"`
	GuardianDID  string `json:"guardian_did"`
	Role         string `json:"role"`
	Status       string `json:"status,omitempty"`
	GuardianVCID string `json:"guardian_vc_id,omitempty"`
}

// InitiateResponse lists guardians the holder should contact for shares.
type InitiateResponse struct {
	Session     Session       `json:"session"`
	Shareholders []Shareholder `json:"shareholders"`
}

// CompleteRequest proves client-side reconstruction without sending the secret.
type CompleteRequest struct {
	SessionID         string `json:"session_id"`
	RootKeyCommitment string `json:"root_key_commitment"`
}

func ValidatePolicy(threshold, total int) error {
	if threshold < 2 || total < threshold || total > 255 {
		return fmt.Errorf("%w: threshold_m=%d total_n=%d", ErrInvalidPolicy, threshold, total)
	}
	return nil
}

func ValidateShareholderInput(in ShareholderInput) error {
	if in.ShareIndex < 1 || in.ShareIndex > 255 {
		return fmt.Errorf("%w: share_index out of range", ErrInvalidShareholder)
	}
	if in.GuardianDID == "" {
		return fmt.Errorf("%w: guardian_did required", ErrInvalidShareholder)
	}
	switch in.Role {
	case RoleDevice, RoleContact, RoleOrg:
	default:
		return fmt.Errorf("%w: role must be device, contact, or org", ErrInvalidShareholder)
	}
	status := in.Status
	if status == "" {
		status = StatusPending
	}
	switch status {
	case StatusPending, StatusActive, StatusRevoked:
	default:
		return fmt.Errorf("%w: invalid status", ErrInvalidShareholder)
	}
	return nil
}

// RejectSecretFields returns ErrForbiddenField if a JSON body includes honeypot key material fields.
func RejectSecretFields(raw map[string]interface{}) error {
	forbidden := []string{
		"share", "share_bytes", "shares", "recovery_secret", "secret",
		"passport_root_key", "root_key", "share_material",
	}
	for _, key := range forbidden {
		if _, ok := raw[key]; ok {
			return fmt.Errorf("%w: field %q", ErrForbiddenField, key)
		}
	}
	return nil
}
