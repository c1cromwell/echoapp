package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// CORSConfig defines the CORS policy configuration
type CORSPolicy struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string
	MaxAge         int
}

// NewIOSCORSPolicy creates a CORS policy that only allows the iOS application
func NewIOSCORSPolicy(iosAppOrigin string) *CORSPolicy {
	return &CORSPolicy{
		AllowedOrigins: []string{iosAppOrigin},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Request-ID",
			"X-iOS-App-Version",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
		},
		MaxAge: 3600,
	}
}

// CORSMiddleware enforces CORS policy and rejects non-allowed origins with 403
// This middleware strictly validates origins against the allowed list
func CORSMiddleware(policy *CORSPolicy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				// Check if origin is in the allowed list
				allowed := false
				for _, allowedOrigin := range policy.AllowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						break
					}
				}

				// Reject non-allowed origins with 403 Forbidden
				if !allowed {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)

					// Write error response
					errResponse := map[string]interface{}{
						"code":        "CORS_POLICY_VIOLATION",
						"message":     "Origin not allowed by CORS policy",
						"origin":      origin,
						"request_id":  getRequestID(r),
						"status_code": http.StatusForbidden,
					}

					json.NewEncoder(w).Encode(errResponse)
					return
				}

				// Set CORS headers for allowed origins
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(policy.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(policy.AllowedHeaders, ", "))
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(policy.ExposedHeaders, ", "))
				w.Header().Set("Access-Control-Max-Age", "3600")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
