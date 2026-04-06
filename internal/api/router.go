// Package api provides the shared HTTP router, middleware, and handlers
// used by both the production server and integration tests.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const (
	// ContextKeyUserID is the context key for the authenticated user ID.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyRequestID is the context key for the request ID.
	ContextKeyRequestID contextKey = "request_id"
)

// APIError represents a standardized error response.
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id"`
	Timestamp  string `json:"timestamp"`
	StatusCode int    `json:"status_code"`
}

// HealthCheckResponse represents health status.
type HealthCheckResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	RequestID string `json:"request_id"`
}

// Router holds the HTTP handler and configuration for the Echo API.
type Router struct {
	AllowedOrigins  []string
	StartTime       time.Time
	TokenValidator  func(token string) bool
	UserIDExtractor func(token string) string
	WSHub           *Hub        // WebSocket hub for real-time messaging
	V3              *V3Handlers // V3 API handlers (blueprint services)
}

// NewRouter creates a Router with sensible defaults.
// It also creates and starts a WebSocket hub.
func NewRouter(allowedOrigins []string) *Router {
	hub := NewHub()
	go hub.Run()

	return &Router{
		AllowedOrigins: allowedOrigins,
		StartTime:      time.Now(),
		TokenValidator: func(token string) bool {
			return len(token) > 0
		},
		UserIDExtractor: func(token string) string {
			return "user-" + token[:min(8, len(token))]
		},
		WSHub: hub,
	}
}

// Handler returns the fully wrapped http.Handler with all middleware applied.
func (rt *Router) Handler() http.Handler {
	core := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/health":
			rt.handleHealth(w, r)
		case r.URL.Path == "/ws":
			// WebSocket upgrade — handled before auth middleware wraps the response
			ServeWS(rt.WSHub, rt.UserIDExtractor, w, r)
		case strings.HasPrefix(r.URL.Path, "/v1/"):
			rt.handleV1(w, r)
		case strings.HasPrefix(r.URL.Path, "/v2/"):
			rt.handleV2(w, r)
		case strings.HasPrefix(r.URL.Path, "/v3/"):
			rt.handleV3(w, r)
		default:
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
		}
	})

	// Middleware chain: CORS -> Auth -> RequestID -> core
	return rt.corsMiddleware(rt.authMiddleware(rt.requestIDMiddleware(core)))
}

// --- Middleware ---

func (rt *Router) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		r.Header.Set("X-Request-ID", requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func (rt *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check and WebSocket don't go through REST auth middleware
		if r.URL.Path == "/health" || r.URL.Path == "/ws" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteError(w, http.StatusUnauthorized, "MISSING_AUTH", "Authorization header required", r.Header.Get("X-Request-ID"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteError(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization must be Bearer token", r.Header.Get("X-Request-ID"))
			return
		}

		token := parts[1]
		if !rt.TokenValidator(token) {
			WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", r.Header.Get("X-Request-ID"))
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, rt.UserIDExtractor(token))
		ctx = context.WithValue(ctx, ContextKeyRequestID, r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (rt *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" {
			allowed := false
			for _, allowedOrigin := range rt.AllowedOrigins {
				if allowedOrigin == "*" || origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if !allowed {
				WriteError(w, http.StatusForbidden, "CORS_DENIED", "Origin not allowed", r.Header.Get("X-Request-ID"))
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Route handlers ---

func (rt *Router) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, HealthCheckResponse{
		Status:    "operational",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    time.Since(rt.StartTime).String(),
		RequestID: r.Header.Get("X-Request-ID"),
	})
}

func (rt *Router) handleV1(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/users":
		rt.v1GetUsers(w, r)
	case "/v1/users/profile":
		rt.v1GetProfile(w, r)
	default:
		WriteError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) v1GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "user1", "name": "Alice"},
			{"id": "user2", "name": "Bob"},
		},
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (rt *Router) v1GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	userID := r.Context().Value(ContextKeyUserID)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":    userID,
		"email":      "user@example.com",
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (rt *Router) handleV2(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v2/users":
		rt.v2GetUsers(w, r)
	case "/v2/users/profile":
		rt.v2GetProfile(w, r)
	default:
		WriteError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
	}
}

// handleV3 delegates to the V3Handlers which connect to real backend services.
func (rt *Router) handleV3(w http.ResponseWriter, r *http.Request) {
	if rt.V3 == nil {
		WriteError(w, http.StatusServiceUnavailable, "V3_NOT_CONFIGURED", "V3 API services not initialized", r.Header.Get("X-Request-ID"))
		return
	}

	// Create a ServeMux and register V3 routes
	mux := http.NewServeMux()
	rt.V3.RegisterV3Routes(mux)
	mux.ServeHTTP(w, r)
}

func (rt *Router) v2GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "user1", "name": "Alice", "status": "active", "created_at": "2025-01-01T00:00:00Z"},
			{"id": "user2", "name": "Bob", "status": "active", "created_at": "2025-01-02T00:00:00Z"},
		},
		"pagination": map[string]interface{}{
			"total": 2,
			"page":  1,
			"limit": 10,
		},
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (rt *Router) v2GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	userID := r.Context().Value(ContextKeyUserID)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":    userID,
		"email":      "user@example.com",
		"phone":      "+1-555-0100",
		"verified":   true,
		"last_login": time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339),
		"created_at": "2024-01-01T00:00:00Z",
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// --- Helpers ---

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a standardized JSON error response.
func WriteError(w http.ResponseWriter, statusCode int, errorCode, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIError{
		Code:       errorCode,
		Message:    message,
		RequestID:  requestID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	})
}
