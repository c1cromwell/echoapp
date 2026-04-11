package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id"`
	Timestamp  string `json:"timestamp"`
	StatusCode int    `json:"status_code"`
}

// RequestMetadata holds request context data
type RequestMetadata struct {
	RequestID string
	UserID    string
	StartTime time.Time
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string
	MaxAge         int
}

// NewCORSConfig creates a new CORS configuration
func NewCORSConfig(allowedOrigins []string) *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposedHeaders: []string{"X-Request-ID", "X-Rate-Limit-Remaining"},
		MaxAge:         3600,
	}
}

// RequestIDMiddleware generates and tracks request IDs
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		r.Header.Set("X-Request-ID", requestID)
		w.Header().Set("X-Request-ID", requestID)

		// Add metadata to context
		metadata := RequestMetadata{
			RequestID: requestID,
			StartTime: time.Now(),
		}
		ctx := context.WithValue(r.Context(), "metadata", metadata)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddleware validates Bearer token authentication using passkey verification
// Returns 401 Unauthorized for invalid or missing authentication
// Parameters:
//   - skipPaths: list of URL paths that don't require authentication
func AuthMiddleware(skipPaths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for certain paths
			for _, path := range skipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, http.StatusUnauthorized, "MISSING_AUTH", "Authorization header required", getRequestID(r))
				return
			}

			// Parse Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeErrorResponse(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization must be Bearer token", getRequestID(r))
				return
			}

			token := parts[1]

			// Validate passkey token
			userID, valid := ValidateToken(token)
			if !valid {
				// Return 401 Unauthorized for invalid passkey verification
				writeErrorResponse(w, http.StatusUnauthorized, "INVALID_PASSKEY", "Passkey verification failed", getRequestID(r))
				return
			}

			// Add user ID to context for use in handlers
			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r)
		start := time.Now()

		// Log request
		log.Printf("[%s] %s %s %s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		// Wrap response writer to capture status code
		wrapped := &wrappedResponseWriter{ResponseWriter: w}

		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		log.Printf("[%s] %d %s %dms", requestID, wrapped.statusCode, r.URL.Path, duration.Milliseconds())
	})
}

// wrappedResponseWriter wraps http.ResponseWriter to capture status code
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// PasskeyValidator defines the interface for passkey verification
type PasskeyValidator interface {
	Verify(token string) (userID string, valid bool, err error)
}

// DefaultPasskeyValidator provides a default implementation of passkey validation
type DefaultPasskeyValidator struct {
	// This would typically contain a reference to your passkey store/service
	// For now, it serves as a placeholder for actual implementation
}

// Verify validates a passkey token
// Parameters:
//   - token: the passkey token to validate
//
// Returns:
//   - userID: the user ID associated with the token if valid
//   - valid: whether the token is valid
//   - err: any error that occurred during validation
func (v *DefaultPasskeyValidator) Verify(token string) (string, bool, error) {
	// Token validation rules:
	// 1. Token must not be empty
	// 2. Token must meet minimum length requirement (e.g., 32 characters for typical passkeys)
	// 3. Token format validation (alphanumeric with hyphens/underscores allowed)

	if len(token) == 0 {
		return "", false, nil
	}

	if len(token) < 32 {
		return "", false, nil
	}

	// Validate token format (example: must contain valid characters)
	for _, r := range token {
		if !isValidTokenChar(r) {
			return "", false, nil
		}
	}

	// Production auth is handled by internal/auth.TokenService (ES256 JWT)
	// and internal/auth.PasskeyVerifier (WebAuthn P-256 assertion).
	// This legacy validator rejects all tokens — it exists as a safe fallback.
	return "", false, nil
}

// isValidTokenChar checks if a character is valid in a token
func isValidTokenChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.'
}

// ValidateToken validates a Bearer token using the default validator
func ValidateToken(token string) (string, bool) {
	validator := &DefaultPasskeyValidator{}
	userID, valid, err := validator.Verify(token)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		return "", false
	}
	return userID, valid
}

// getRequestID extracts request ID from context or headers
func getRequestID(r *http.Request) string {
	if metadata, ok := r.Context().Value("metadata").(RequestMetadata); ok {
		return metadata.RequestID
	}
	return r.Header.Get("X-Request-ID")
}

// writeErrorResponse writes a standardized error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResponse := ErrorResponse{
		Code:       errorCode,
		Message:    message,
		RequestID:  requestID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	}

	json.NewEncoder(w).Encode(errResponse)
}
