package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// ErrorResponse represents a standardized error response envelope
type ErrorResponse struct {
	Error      ErrorDetails `json:"error"`
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	RequestID  string       `json:"request_id"`
	Timestamp  string       `json:"timestamp"`
	StatusCode int          `json:"status_code"`
}

// ErrorDetails contains additional error information
type ErrorDetails struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// WriteError writes a standardized error response to the HTTP response writer
// Parameters:
//   - w: http.ResponseWriter to write the response to
//   - statusCode: HTTP status code (e.g., http.StatusBadRequest, http.StatusNotFound)
//   - code: Application-specific error code (e.g., "INVALID_REQUEST", "NOT_FOUND")
//   - message: Human-readable error message
//   - description: Optional detailed error description
//   - requestID: Request identifier for tracking
func WriteError(w http.ResponseWriter, statusCode int, code, message, description, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: ErrorDetails{
			Code:        code,
			Message:     message,
			Description: description,
		},
		Code:       code,
		Message:    message,
		RequestID:  requestID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteErrorWithDefaults writes an error response with default message for common HTTP status codes
// Parameters:
//   - w: http.ResponseWriter to write the response to
//   - statusCode: HTTP status code (e.g., http.StatusBadRequest, http.StatusNotFound)
//   - code: Application-specific error code
//   - description: Optional detailed error description
//   - requestID: Request identifier for tracking
//
// Default messages for common status codes:
//   - 400 Bad Request: "Invalid request"
//   - 401 Unauthorized: "Authentication required"
//   - 403 Forbidden: "Access denied"
//   - 404 Not Found: "Resource not found"
//   - 409 Conflict: "Resource conflict"
//   - 422 Unprocessable Entity: "Validation failed"
//   - 429 Too Many Requests: "Rate limit exceeded"
//   - 500 Internal Server Error: "Internal server error"
//   - 503 Service Unavailable: "Service unavailable"
func WriteErrorWithDefaults(w http.ResponseWriter, statusCode int, code, description, requestID string) {
	message := getDefaultErrorMessage(statusCode)
	WriteError(w, statusCode, code, message, description, requestID)
}

// getDefaultErrorMessage returns default error message for common HTTP status codes
func getDefaultErrorMessage(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Invalid request"
	case http.StatusUnauthorized:
		return "Authentication required"
	case http.StatusForbidden:
		return "Access denied"
	case http.StatusNotFound:
		return "Resource not found"
	case http.StatusConflict:
		return "Resource conflict"
	case http.StatusUnprocessableEntity:
		return "Validation failed"
	case http.StatusTooManyRequests:
		return "Rate limit exceeded"
	case http.StatusInternalServerError:
		return "Internal server error"
	case http.StatusServiceUnavailable:
		return "Service unavailable"
	case http.StatusMethodNotAllowed:
		return "Method not allowed"
	default:
		return "An error occurred"
	}
}

// Common Error Codes (constants for consistency)
const (
	// Client Errors
	ErrInvalidRequest    = "INVALID_REQUEST"
	ErrInvalidPayload    = "INVALID_PAYLOAD"
	ErrMissingField      = "MISSING_FIELD"
	ErrMethodNotAllowed  = "METHOD_NOT_ALLOWED"
	ErrNotFound          = "NOT_FOUND"
	ErrUnauthorized      = "UNAUTHORIZED"
	ErrForbidden         = "FORBIDDEN"
	ErrConflict          = "CONFLICT"
	ErrValidationFailed  = "VALIDATION_FAILED"
	ErrResourceConflict  = "RESOURCE_CONFLICT"
	ErrRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrTooManyRequests   = "TOO_MANY_REQUESTS"

	// Server Errors
	ErrInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrTimeout            = "REQUEST_TIMEOUT"
	ErrDatabaseError      = "DATABASE_ERROR"
)
