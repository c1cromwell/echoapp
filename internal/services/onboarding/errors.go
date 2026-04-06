package onboarding

import "errors"

var (
	// Session errors
	ErrSessionNotFound  = errors.New("onboarding session not found")
	ErrSessionExpired   = errors.New("onboarding session expired")
	ErrSessionCompleted = errors.New("onboarding session already completed")
	ErrInvalidStep      = errors.New("invalid onboarding step")
	ErrStepNotReady     = errors.New("previous required steps not completed")

	// Phone verification errors
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	ErrOTPExpired         = errors.New("verification code expired")
	ErrOTPInvalid         = errors.New("invalid verification code")
	ErrOTPRateLimited     = errors.New("too many verification attempts, try again later")
	ErrOTPMaxAttempts     = errors.New("maximum verification attempts exceeded")
	ErrOTPAlreadySent     = errors.New("verification code already sent, wait before resending")

	// Passkey errors
	ErrPasskeyCreationFailed = errors.New("passkey creation failed")
	ErrPasskeyInvalidData    = errors.New("invalid passkey credential data")
	ErrPasskeyAlreadyExists  = errors.New("passkey already registered for this session")

	// Recovery errors
	ErrRecoveryMethodInvalid = errors.New("invalid recovery method type")
	ErrRecoveryEmailInvalid  = errors.New("invalid recovery email address")
	ErrRecoveryEmailNotVerified = errors.New("recovery email not verified")
	ErrRecoveryWalletInvalid = errors.New("invalid wallet address")
	ErrRecoveryContactInvalid = errors.New("invalid trusted contact")

	// Profile errors
	ErrDisplayNameRequired = errors.New("display name is required")
	ErrDisplayNameTooLong  = errors.New("display name exceeds maximum length")
	ErrUsernameTaken       = errors.New("username is already taken")
	ErrUsernameInvalid     = errors.New("username contains invalid characters")
	ErrBioTooLong          = errors.New("bio exceeds 150 character limit")
)
