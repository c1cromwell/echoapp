package auth

import (
	"time"
)

// StepUpAction defines actions that require elevated authentication.
type StepUpAction string

const (
	StepUpChangePhone    StepUpAction = "change_phone"
	StepUpRevokeDevice   StepUpAction = "revoke_device"
	StepUpLargePayment   StepUpAction = "large_payment"
	StepUpExportRecovery StepUpAction = "export_recovery"
	StepUpDeleteAccount  StepUpAction = "delete_account"
)

// StepUpMethod defines how to verify for step-up.
type StepUpMethod string

const (
	StepUpMethodPasskey    StepUpMethod = "passkey"
	StepUpMethodOTP        StepUpMethod = "otp"
	StepUpMethodPasskeyOTP StepUpMethod = "passkey+otp"
)

// StepUpRequirement defines what's needed for a given action.
type StepUpRequirement struct {
	Action         StepUpAction
	RequiredMethod StepUpMethod
	TokenTTL       time.Duration
}

// StepUpService manages step-up authentication requirements and verification.
type StepUpService struct {
	requirements map[StepUpAction]*StepUpRequirement
}

// NewStepUpService creates a step-up service with default requirements.
func NewStepUpService() *StepUpService {
	s := &StepUpService{
		requirements: make(map[StepUpAction]*StepUpRequirement),
	}

	// Configure requirements per the spec
	s.requirements[StepUpChangePhone] = &StepUpRequirement{
		Action:         StepUpChangePhone,
		RequiredMethod: StepUpMethodOTP,
		TokenTTL:       5 * time.Minute,
	}
	s.requirements[StepUpRevokeDevice] = &StepUpRequirement{
		Action:         StepUpRevokeDevice,
		RequiredMethod: StepUpMethodPasskey,
		TokenTTL:       5 * time.Minute,
	}
	s.requirements[StepUpLargePayment] = &StepUpRequirement{
		Action:         StepUpLargePayment,
		RequiredMethod: StepUpMethodPasskey,
		TokenTTL:       5 * time.Minute,
	}
	s.requirements[StepUpExportRecovery] = &StepUpRequirement{
		Action:         StepUpExportRecovery,
		RequiredMethod: StepUpMethodPasskeyOTP,
		TokenTTL:       2 * time.Minute,
	}
	s.requirements[StepUpDeleteAccount] = &StepUpRequirement{
		Action:         StepUpDeleteAccount,
		RequiredMethod: StepUpMethodPasskeyOTP,
		TokenTTL:       0, // Queued with 24hr cooling, not immediate
	}

	return s
}

// GetRequirement returns the step-up requirement for an action.
func (s *StepUpService) GetRequirement(action StepUpAction) *StepUpRequirement {
	return s.requirements[action]
}

// RequiresStepUp checks if an action requires elevated authentication.
func (s *StepUpService) RequiresStepUp(action string) bool {
	_, exists := s.requirements[StepUpAction(action)]
	return exists
}

// ValidateMethod checks that the provided method satisfies the requirement.
func (s *StepUpService) ValidateMethod(action StepUpAction, method StepUpMethod) bool {
	req, exists := s.requirements[action]
	if !exists {
		return false
	}

	// passkey+otp satisfies any requirement
	if method == StepUpMethodPasskeyOTP {
		return true
	}

	return req.RequiredMethod == method
}

// IsElevatedForAction checks if token claims are elevated for a specific action.
func IsElevatedForAction(claims *TokenClaims, action string) bool {
	if claims == nil {
		return false
	}
	return claims.Elevated && claims.ElevatedAction == action
}
