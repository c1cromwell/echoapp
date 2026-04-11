package auth

import (
	"time"
)

// --- User ---

// UserStatus represents the user account lifecycle state.
type UserStatus string

const (
	UserStatusPendingPasskey UserStatus = "pending_passkey"
	UserStatusActive         UserStatus = "active"
	UserStatusSuspended      UserStatus = "suspended"
	UserStatusDeleted        UserStatus = "deleted"
)

// User is the core user entity stored in the users table.
type User struct {
	ID          string     `json:"id"`
	DID         string     `json:"did"`
	PhoneHash   string     `json:"-"` // Never expose phone hash
	DisplayName string     `json:"display_name,omitempty"`
	Username    string     `json:"username,omitempty"`
	Status      UserStatus `json:"status"`
	TrustScore  int        `json:"trust_score"`
	TrustTier   int        `json:"trust_tier"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// --- Device ---

// DeviceInfo is sent by the client in the X-Device-Info header.
type DeviceInfo struct {
	DeviceID        string `json:"device_id"         validate:"required"`
	Platform        string `json:"platform"          validate:"required,oneof=ios"`
	OSVersion       string `json:"os_version"        validate:"required"`
	AppVersion      string `json:"app_version"       validate:"required"`
	Model           string `json:"model"             validate:"required"`
	Locale          string `json:"locale"`
	Timezone        string `json:"timezone"`
	SecureEnclave   bool   `json:"secure_enclave"`
	BiometricType   string `json:"biometric_type"    validate:"oneof=face_id touch_id none"`
	JailbreakStatus bool   `json:"jailbreak_status"`
	PushToken       string `json:"push_token,omitempty"`
}

// DeviceRecord is the persisted device entity.
type DeviceRecord struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	DeviceHash      string    `json:"device_hash"`
	FriendlyName    string    `json:"friendly_name"`
	Platform        string    `json:"platform"`
	OSVersion       string    `json:"os_version"`
	LastIP          string    `json:"last_ip"`
	LastLocation    string    `json:"last_location,omitempty"`
	LastActiveAt    time.Time `json:"last_active_at"`
	CreatedAt       time.Time `json:"created_at"`
	CredentialID    string    `json:"credential_id,omitempty"`
	IsCurrentDevice bool      `json:"is_current_device"`
}

// --- Credentials (Passkeys) ---

// CredentialRecord represents a stored WebAuthn passkey credential.
type CredentialRecord struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	CredentialID string     `json:"credential_id"`
	PublicKey    []byte     `json:"-"`
	SignCount    int64      `json:"sign_count"`
	DeviceID     string     `json:"device_id"`
	AAGUID       string     `json:"aaguid,omitempty"`
	FriendlyName string     `json:"friendly_name,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

// --- Sessions & Refresh Tokens ---

// RefreshTokenStatus tracks the lifecycle of a refresh token.
type RefreshTokenStatus string

const (
	RefreshTokenActive  RefreshTokenStatus = "active"
	RefreshTokenUsed    RefreshTokenStatus = "used"
	RefreshTokenRevoked RefreshTokenStatus = "revoked"
)

// RefreshTokenRecord is a stored refresh token.
type RefreshTokenRecord struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	TokenHash string             `json:"-"`
	DeviceID  string             `json:"device_id"`
	Status    RefreshTokenStatus `json:"status"`
	ExpiresAt time.Time          `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	UsedAt    *time.Time         `json:"used_at,omitempty"`
}

// --- Audit Log ---

// AuditEventType categorizes authentication events.
type AuditEventType string

const (
	AuditEventLogin    AuditEventType = "login"
	AuditEventRegister AuditEventType = "register"
	AuditEventRefresh  AuditEventType = "refresh"
	AuditEventLogout   AuditEventType = "logout"
	AuditEventStepUp   AuditEventType = "step_up"
	AuditEventRecovery AuditEventType = "recovery"
)

// AuditResult is the outcome of an auth event.
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailed  AuditResult = "failed"
	AuditResultBlocked AuditResult = "blocked"
)

// AuditLogEntry records an authentication event for security monitoring.
type AuditLogEntry struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id,omitempty"`
	EventType  AuditEventType         `json:"event_type"`
	Result     AuditResult            `json:"result"`
	IPAddress  string                 `json:"ip_address"`
	DeviceID   string                 `json:"device_id,omitempty"`
	DeviceInfo *DeviceInfo            `json:"device_info,omitempty"`
	ErrorCode  string                 `json:"error_code,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// --- API Request/Response Types ---

// AuthResponse is the standard authentication response.
type AuthResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	ExpiresAt        string `json:"expires_at"`
	User             *User  `json:"user,omitempty"`
	PasskeyChallenge string `json:"passkey_challenge,omitempty"`
}

// PhoneRegistrationRequest starts the OTP flow.
type PhoneRegistrationRequest struct {
	PhoneNumber string     `json:"phone_number" validate:"required"`
	CountryCode string     `json:"country_code" validate:"required"`
	DeviceInfo  DeviceInfo `json:"device_info"`
}

// PhoneRegistrationResponse is returned after OTP is sent.
type PhoneRegistrationResponse struct {
	VerificationID string `json:"verification_id"`
	ExpiresAt      string `json:"expires_at"`
	RetryAfter     int    `json:"retry_after"`
}

// OTPVerifyRequest verifies an OTP code.
type OTPVerifyRequest struct {
	VerificationID string     `json:"verification_id" validate:"required"`
	Code           string     `json:"code" validate:"required,len=6"`
	DeviceInfo     DeviceInfo `json:"device_info"`
}

// PasskeyRegistrationRequest registers a new passkey.
type PasskeyRegistrationRequest struct {
	Challenge           string              `json:"challenge"`
	AttestationResponse AttestationResponse `json:"attestation_response"`
	DeviceInfo          DeviceInfo          `json:"device_info"`
}

// AttestationResponse is the WebAuthn attestation from the client.
type AttestationResponse struct {
	ID       string                    `json:"id"`
	RawID    string                    `json:"raw_id"`
	Response AttestationResponseDetail `json:"response"`
	Type     string                    `json:"type"`
}

// AttestationResponseDetail contains the attestation data.
type AttestationResponseDetail struct {
	ClientDataJSON    string `json:"client_data_json"`
	AttestationObject string `json:"attestation_object"`
}

// LoginRequest handles passkey or DID signature login.
type LoginRequest struct {
	AuthType   string           `json:"auth_type" validate:"required,oneof=passkey did_signature"`
	Credential *LoginCredential `json:"credential,omitempty"`
	DID        string           `json:"did,omitempty"`
	Signature  string           `json:"signature,omitempty"`
	Timestamp  string           `json:"timestamp,omitempty"`
	Nonce      string           `json:"nonce,omitempty"`
	DeviceInfo DeviceInfo       `json:"device_info"`
}

// LoginCredential is the WebAuthn assertion from the client.
type LoginCredential struct {
	ID       string                `json:"id"`
	RawID    string                `json:"raw_id"`
	Response AssertionResponseData `json:"response"`
	Type     string                `json:"type"`
}

// AssertionResponseData contains the assertion data.
type AssertionResponseData struct {
	ClientDataJSON    string `json:"client_data_json"`
	AuthenticatorData string `json:"authenticator_data"`
	Signature         string `json:"signature"`
}

// LoginChallengeResponse is returned for pre-login challenge.
type LoginChallengeResponse struct {
	Challenge string `json:"challenge"`
	Timeout   int    `json:"timeout"`
	RPID      string `json:"rp_id"`
}

// RefreshRequest rotates the refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest invalidates sessions.
type LogoutRequest struct {
	AllDevices bool `json:"all_devices"`
}

// StepUpRequest elevates a session for sensitive operations.
type StepUpRequest struct {
	Method     string           `json:"method" validate:"required,oneof=passkey otp passkey+otp"`
	Credential *LoginCredential `json:"credential,omitempty"`
	OTPCode    string           `json:"otp_code,omitempty"`
	Action     string           `json:"action" validate:"required"`
}

// StepUpResponse returns an elevated token.
type StepUpResponse struct {
	ElevatedToken string `json:"elevated_token"`
	ExpiresAt     string `json:"expires_at"`
	Action        string `json:"action"`
}

// RecoveryInitiateRequest starts account recovery.
type RecoveryInitiateRequest struct {
	RecoveryMethod string     `json:"recovery_method" validate:"required,oneof=recovery_phrase trusted_contacts phone"`
	Identifier     string     `json:"identifier" validate:"required"`
	DeviceInfo     DeviceInfo `json:"device_info"`
}

// RecoveryInitiateResponse returns the recovery session info.
type RecoveryInitiateResponse struct {
	RecoverySessionID string   `json:"recovery_session_id"`
	RequiredSteps     []string `json:"required_steps"`
	ExpiresAt         string   `json:"expires_at"`
}

// RecoveryVerifyRequest completes account recovery.
type RecoveryVerifyRequest struct {
	RecoverySessionID string `json:"recovery_session_id" validate:"required"`
	Method            string `json:"method" validate:"required"`
	Proof             string `json:"proof,omitempty"`
}

// --- JWT Claims ---

// TokenClaims are the custom claims in an ECHO access token.
type TokenClaims struct {
	Subject        string `json:"sub"`
	IssuedAt       int64  `json:"iat"`
	ExpiresAt      int64  `json:"exp"`
	TokenID        string `json:"jti"`
	DeviceID       string `json:"device_id"`
	TrustTier      int    `json:"trust_tier"`
	Scope          string `json:"scope"`
	Elevated       bool   `json:"elevated"`
	ElevatedAction string `json:"elevated_action,omitempty"`
}

// --- Error Types ---

// AuthErrorCode is a structured error code for the auth system.
type AuthErrorCode string

const (
	ErrCodeInvalidPhone    AuthErrorCode = "AUTH_001"
	ErrCodeOTPRateLimit    AuthErrorCode = "AUTH_002"
	ErrCodeInvalidOTP      AuthErrorCode = "AUTH_003"
	ErrCodePasskeyFailed   AuthErrorCode = "AUTH_004"
	ErrCodeTokenExpired    AuthErrorCode = "AUTH_005"
	ErrCodeRefreshInvalid  AuthErrorCode = "AUTH_006"
	ErrCodeUnknownDevice   AuthErrorCode = "AUTH_007"
	ErrCodeStepUpRequired  AuthErrorCode = "AUTH_008"
	ErrCodeAccountLocked   AuthErrorCode = "AUTH_009"
	ErrCodeDeviceIntegrity AuthErrorCode = "AUTH_010"
	ErrCodeRecoveryInvalid AuthErrorCode = "AUTH_011"
	ErrCodeGlobalRateLimit AuthErrorCode = "AUTH_012"
)

// AuthError is a structured authentication error.
type AuthError struct {
	Code       AuthErrorCode `json:"code"`
	Message    string        `json:"message"`
	HTTPStatus int           `json:"-"`
	RetryAfter *int          `json:"retry_after,omitempty"`
}

func (e *AuthError) Error() string {
	return string(e.Code) + ": " + e.Message
}

// UserFacingMessage returns the safe message shown to users.
func (c AuthErrorCode) UserFacingMessage() string {
	switch c {
	case ErrCodeInvalidPhone:
		return "Please enter a valid phone number."
	case ErrCodeOTPRateLimit:
		return "Too many attempts. Please try again later."
	case ErrCodeInvalidOTP:
		return "That code is incorrect. Please try again."
	case ErrCodePasskeyFailed:
		return "Authentication failed. Please try again."
	case ErrCodeTokenExpired:
		return "Your session has expired. Please sign in."
	case ErrCodeRefreshInvalid:
		return "Please sign in again."
	case ErrCodeUnknownDevice:
		return "New device detected. Verify your identity."
	case ErrCodeStepUpRequired:
		return "Additional verification required."
	case ErrCodeAccountLocked:
		return "Account locked. Try again in 1 hour."
	case ErrCodeDeviceIntegrity:
		return "Device verification failed."
	case ErrCodeRecoveryInvalid:
		return "Recovery session expired. Please restart."
	case ErrCodeGlobalRateLimit:
		return "Too many requests. Please slow down."
	default:
		return "An error occurred. Please try again."
	}
}

// NewAuthError creates a structured auth error.
func NewAuthError(code AuthErrorCode, httpStatus int) *AuthError {
	return &AuthError{
		Code:       code,
		Message:    code.UserFacingMessage(),
		HTTPStatus: httpStatus,
	}
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error *AuthError `json:"error"`
}
