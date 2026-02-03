package credentials

import (
	"fmt"
	"time"
)

// CredentialErrorCode represents credential-specific error codes
type CredentialErrorCode string

const (
	ErrCodeInvalidCredential     CredentialErrorCode = "INVALID_CREDENTIAL"
	ErrCodeInvalidFormat         CredentialErrorCode = "INVALID_FORMAT"
	ErrCodeInvalidSignature      CredentialErrorCode = "INVALID_SIGNATURE"
	ErrCodeInvalidIssuer         CredentialErrorCode = "INVALID_ISSUER"
	ErrCodeExpiredCredential     CredentialErrorCode = "EXPIRED_CREDENTIAL"
	ErrCodeRevokedCredential     CredentialErrorCode = "REVOKED_CREDENTIAL"
	ErrCodeCredentialNotFound    CredentialErrorCode = "CREDENTIAL_NOT_FOUND"
	ErrCodeIssuanceFailed        CredentialErrorCode = "ISSUANCE_FAILED"
	ErrCodeVerificationFailed    CredentialErrorCode = "VERIFICATION_FAILED"
	ErrCodeUnsupportedFormat     CredentialErrorCode = "UNSUPPORTED_FORMAT"
	ErrCodeInvalidSubject        CredentialErrorCode = "INVALID_SUBJECT"
	ErrCodeStorageFailed         CredentialErrorCode = "STORAGE_FAILED"
	ErrCodeBlockchainError       CredentialErrorCode = "BLOCKCHAIN_ERROR"
	ErrCodeRevocationCheckFailed CredentialErrorCode = "REVOCATION_CHECK_FAILED"
	ErrCodeInvalidProof          CredentialErrorCode = "INVALID_PROOF"
	ErrCodeMissingContext        CredentialErrorCode = "MISSING_CONTEXT"
	ErrCodeMissingIssuer         CredentialErrorCode = "MISSING_ISSUER"
	ErrCodeMissingSubject        CredentialErrorCode = "MISSING_SUBJECT"
	ErrCodeInvalidIssuanceDate   CredentialErrorCode = "INVALID_ISSUANCE_DATE"
	ErrCodeInvalidExpirationDate CredentialErrorCode = "INVALID_EXPIRATION_DATE"
	ErrCodeTimeoutError          CredentialErrorCode = "TIMEOUT_ERROR"
	ErrCodeUnauthorized          CredentialErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidRequest        CredentialErrorCode = "INVALID_REQUEST"
)

// CredentialError represents a credential-related error
type CredentialError struct {
	Code      CredentialErrorCode
	Message   string
	Details   string
	Timestamp int64
}

// Error implements the error interface
func (e *CredentialError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewCredentialError creates a new credential error
func NewCredentialError(code CredentialErrorCode, message string) *CredentialError {
	return &CredentialError{
		Code:      code,
		Message:   message,
		Timestamp: now().Unix(),
	}
}

// NewCredentialErrorWithDetails creates a credential error with details
func NewCredentialErrorWithDetails(code CredentialErrorCode, message, details string) *CredentialError {
	return &CredentialError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: now().Unix(),
	}
}

// IsCredentialError checks if an error is a CredentialError
func IsCredentialError(err error) bool {
	_, ok := err.(*CredentialError)
	return ok
}

// GetCredentialErrorCode gets error code if error is CredentialError
func GetCredentialErrorCode(err error) CredentialErrorCode {
	if credErr, ok := err.(*CredentialError); ok {
		return credErr.Code
	}
	return ""
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationErrors represents collection of validation errors
type ValidationErrors []ValidationError

// Error implements error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation errors"
	}
	return fmt.Sprintf("validation error: %s", v[0].Message)
}

// Add appends validation error
func (v *ValidationErrors) Add(field, message, code string) {
	*v = append(*v, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// HasErrors checks if any validation errors exist
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// OIDC4VC Error Response Codes
type OIDC4VCErrorCode string

const (
	OIDCErrInvalidRequest              OIDC4VCErrorCode = "invalid_request"
	OIDCErrInvalidClient               OIDC4VCErrorCode = "invalid_client"
	OIDCErrInvalidGrant                OIDC4VCErrorCode = "invalid_grant"
	OIDCErrUnauthorizedClient          OIDC4VCErrorCode = "unauthorized_client"
	OIDCErrUnsupportedGrantType        OIDC4VCErrorCode = "unsupported_grant_type"
	OIDCErrInvalidScope                OIDC4VCErrorCode = "invalid_scope"
	OIDCErrServerError                 OIDC4VCErrorCode = "server_error"
	OIDCErrTemporarilyUnavailable      OIDC4VCErrorCode = "temporarily_unavailable"
	OIDCErrInvalidCredentialRequest    OIDC4VCErrorCode = "invalid_credential_request"
	OIDCErrUnsupportedCredentialType   OIDC4VCErrorCode = "unsupported_credential_type"
	OIDCErrUnsupportedCredentialFormat OIDC4VCErrorCode = "unsupported_credential_format"
	OIDCErrUnsupportedProofType        OIDC4VCErrorCode = "unsupported_proof_type"
	OIDCErrInvalidProof                OIDC4VCErrorCode = "invalid_proof"
	OIDCErrInvalidEncryption           OIDC4VCErrorCode = "invalid_encryption_parameters"
	OIDCErrTxCodeRequired              OIDC4VCErrorCode = "tx_code_required"
)

// OIDC4VCErrorResponse represents OIDC4VC error response
type OIDC4VCErrorResponse struct {
	Error            OIDC4VCErrorCode `json:"error"`
	ErrorDescription string           `json:"error_description,omitempty"`
	ErrorURI         string           `json:"error_uri,omitempty"`
	ErrorCode        int              `json:"error_code,omitempty"`
}

// NewOIDC4VCErrorResponse creates OIDC4VC error response
func NewOIDC4VCErrorResponse(code OIDC4VCErrorCode, description string) *OIDC4VCErrorResponse {
	return &OIDC4VCErrorResponse{
		Error:            code,
		ErrorDescription: description,
	}
}

// Helper function to get current time (for testability)
var now = time.Now
