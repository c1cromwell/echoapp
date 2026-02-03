package did

import (
	"errors"
	"fmt"
)

// DIDError represents a DID-specific error
type DIDError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *DIDError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *DIDError) Unwrap() error {
	return e.Err
}

// NewDIDError creates a new DID error
func NewDIDError(code, message string, err error) *DIDError {
	return &DIDError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// DID Error Codes and Instances
const (
	ErrCodeInvalidDID             = "INVALID_DID"
	ErrCodeDIDNotFound            = "DID_NOT_FOUND"
	ErrCodeDIDAlreadyExists       = "DID_ALREADY_EXISTS"
	ErrCodeGenerationFailed       = "GENERATION_FAILED"
	ErrCodeAnchoringFailed        = "ANCHORING_FAILED"
	ErrCodeResolutionFailed       = "RESOLUTION_FAILED"
	ErrCodeInvalidDocument        = "INVALID_DOCUMENT"
	ErrCodeDocumentNotAnchored    = "DOCUMENT_NOT_ANCHORED"
	ErrCodeDeviceNotFound         = "DEVICE_NOT_FOUND"
	ErrCodeDeviceAlreadyExists    = "DEVICE_ALREADY_EXISTS"
	ErrCodeInvalidPublicKey       = "INVALID_PUBLIC_KEY"
	ErrCodeCryptoFailed           = "CRYPTO_FAILED"
	ErrCodeQRCodeGenerationFailed = "QR_CODE_GENERATION_FAILED"
	ErrCodeDatabaseError          = "DATABASE_ERROR"
	ErrCodeCacheError             = "CACHE_ERROR"
	ErrCodeBlockchainError        = "BLOCKCHAIN_ERROR"
	ErrCodeAtalaPRISMError        = "ATALA_PRISM_ERROR"
	ErrCodeTimeout                = "TIMEOUT"
	ErrCodeInvalidRequest         = "INVALID_REQUEST"
	ErrCodeUnauthorized           = "UNAUTHORIZED"
	ErrCodeInternalError          = "INTERNAL_ERROR"
)

var (
	// DID-specific errors
	ErrInvalidDID             = NewDIDError(ErrCodeInvalidDID, "The provided DID format is invalid", nil)
	ErrDIDNotFound            = NewDIDError(ErrCodeDIDNotFound, "The DID was not found", nil)
	ErrDIDAlreadyExists       = NewDIDError(ErrCodeDIDAlreadyExists, "The DID already exists", nil)
	ErrGenerationFailed       = NewDIDError(ErrCodeGenerationFailed, "Failed to generate DID", nil)
	ErrAnchoringFailed        = NewDIDError(ErrCodeAnchoringFailed, "Failed to anchor DID to blockchain", nil)
	ErrResolutionFailed       = NewDIDError(ErrCodeResolutionFailed, "Failed to resolve DID", nil)
	ErrInvalidDocument        = NewDIDError(ErrCodeInvalidDocument, "The DID document is invalid", nil)
	ErrDocumentNotAnchored    = NewDIDError(ErrCodeDocumentNotAnchored, "The DID document has not been anchored to the blockchain", nil)
	ErrDeviceNotFound         = NewDIDError(ErrCodeDeviceNotFound, "The device was not found", nil)
	ErrDeviceAlreadyExists    = NewDIDError(ErrCodeDeviceAlreadyExists, "The device already exists", nil)
	ErrInvalidPublicKey       = NewDIDError(ErrCodeInvalidPublicKey, "The provided public key is invalid", nil)
	ErrCryptoFailed           = NewDIDError(ErrCodeCryptoFailed, "Cryptographic operation failed", nil)
	ErrQRCodeGenerationFailed = NewDIDError(ErrCodeQRCodeGenerationFailed, "Failed to generate QR code", nil)
	ErrDatabaseError          = NewDIDError(ErrCodeDatabaseError, "Database operation failed", nil)
	ErrCacheError             = NewDIDError(ErrCodeCacheError, "Cache operation failed", nil)
	ErrBlockchainError        = NewDIDError(ErrCodeBlockchainError, "Blockchain operation failed", nil)
	ErrAtalaPRISMError        = NewDIDError(ErrCodeAtalaPRISMError, "Atala PRISM operation failed", nil)
	ErrTimeout                = NewDIDError(ErrCodeTimeout, "Operation timed out", nil)
	ErrInvalidRequest         = NewDIDError(ErrCodeInvalidRequest, "Invalid request", nil)
	ErrUnauthorized           = NewDIDError(ErrCodeUnauthorized, "Unauthorized", nil)
	ErrInternalError          = NewDIDError(ErrCodeInternalError, "Internal server error", nil)
)

// IsDIDError checks if an error is a DIDError with the given code
func IsDIDError(err error, code string) bool {
	var didErr *DIDError
	if errors.As(err, &didErr) {
		return didErr.Code == code
	}
	return false
}

// GetDIDErrorCode extracts the error code from a DIDError
func GetDIDErrorCode(err error) string {
	var didErr *DIDError
	if errors.As(err, &didErr) {
		return didErr.Code
	}
	return ErrCodeInternalError
}

// ValidationError represents validation errors with field information
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors is a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

// Error implements the error interface for ValidationErrors
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation error"
	}
	return fmt.Sprintf("validation error: %d field(s) invalid", len(ve.Errors))
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}
