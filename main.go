package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// APIConfig holds the server configuration
type APIConfig struct {
	Port             string
	TLSCertFile      string
	TLSKeyFile       string
	AllowedOrigins   []string
	RequireAuthToken bool
	ShutdownTimeout  time.Duration
}

// APIError represents a standardized error response
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id"`
	Timestamp  string `json:"timestamp"`
	StatusCode int    `json:"status_code"`
}

// HealthCheckResponse represents health status
type HealthCheckResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	RequestID string `json:"request_id"`
}

// APIServer manages the HTTP server and routing
type APIServer struct {
	config    APIConfig
	server    *http.Server
	startTime time.Time
	mux       *http.ServeMux
}

// RequestContext holds request metadata
type RequestContext struct {
	RequestID string
	UserID    string
	Version   string
}

// NewAPIServer creates a new API server instance
func NewAPIServer(config APIConfig) *APIServer {
	return &APIServer{
		config:    config,
		startTime: time.Now(),
		mux:       http.NewServeMux(),
	}
}

// setupTLS configures TLS 1.3+ settings
func (s *APIServer) setupTLS() *tls.Config {
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
	}
	return tlsConfig
}

// corsMiddleware validates CORS policies
func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" {
			allowed := false
			for _, allowedOrigin := range s.config.AllowedOrigins {
				if allowedOrigin == "*" || origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if !allowed {
				writeError(w, http.StatusForbidden, "CORS_DENIED", "Origin not allowed", r.Header.Get("X-Request-ID"))
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates authentication tokens
func (s *APIServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check endpoint doesn't require auth
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "MISSING_AUTH", "Authorization header required", r.Header.Get("X-Request-ID"))
			return
		}

		// Validate Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization must be Bearer token", r.Header.Get("X-Request-ID"))
			return
		}

		token := parts[1]

		// TODO: Implement passkey verification
		// This is a placeholder - replace with actual passkey validation
		if !validateToken(token) {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", r.Header.Get("X-Request-ID"))
			return
		}

		// Extract user info from token (placeholder)
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", extractUserID(token))
		ctx = context.WithValue(ctx, "request_id", r.Header.Get("X-Request-ID"))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requestIDMiddleware generates request IDs for tracking
func (s *APIServer) requestIDMiddleware(next http.Handler) http.Handler {
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

// healthCheckHandler handles GET /health
func (s *APIServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	uptime := time.Since(s.startTime)
	response := HealthCheckResponse{
		Status:    "operational",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    uptime.String(),
		RequestID: r.Header.Get("X-Request-ID"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// v1Handler serves as a placeholder for v1 endpoint group
func (s *APIServer) v1Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/users":
		s.v1GetUsers(w, r)
	case "/v1/users/profile":
		s.v1GetProfile(w, r)
	default:
		writeError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
	}
}

// v1GetUsers handles GET /v1/users
func (s *APIServer) v1GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "user1", "name": "Alice"},
			{"id": "user2", "name": "Bob"},
		},
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// v1GetProfile handles GET /v1/users/profile
func (s *APIServer) v1GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	userID := r.Context().Value("user_id")
	response := map[string]interface{}{
		"user_id":    userID,
		"email":      "user@example.com",
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// v2Handler serves as a placeholder for v2 endpoint group
func (s *APIServer) v2Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v2/users":
		s.v2GetUsers(w, r)
	case "/v2/users/profile":
		s.v2GetProfile(w, r)
	default:
		writeError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
	}
}

// v2GetUsers handles GET /v2/users (enhanced version)
func (s *APIServer) v2GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	response := map[string]interface{}{
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
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// v2GetProfile handles GET /v2/users/profile (enhanced version)
func (s *APIServer) v2GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	userID := r.Context().Value("user_id")
	response := map[string]interface{}{
		"user_id":    userID,
		"email":      "user@example.com",
		"phone":      "+1-555-0100",
		"verified":   true,
		"last_login": time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339),
		"created_at": "2024-01-01T00:00:00Z",
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes sets up all API routes with middleware
func (s *APIServer) RegisterRoutes() {
	// Wrap all handlers with request ID middleware first, then auth, then CORS
	corsHandler := s.corsMiddleware(s.authMiddleware(s.requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/") {
			s.v1Handler(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/v2/") {
			s.v2Handler(w, r)
		} else if r.URL.Path == "/health" {
			s.healthCheckHandler(w, r)
		} else {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
		}
	}))))

	s.mux.Handle("/", corsHandler)
}

// Start starts the API server
func (s *APIServer) Start() error {
	s.RegisterRoutes()

	listener, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.config.Port, err)
	}

	tlsConfig := s.setupTLS()

	s.server = &http.Server{
		Addr:      ":" + s.config.Port,
		Handler:   s.mux,
		TLSConfig: tlsConfig,
	}

	log.Printf("API Server starting on port %s (TLS 1.3+)", s.config.Port)

	// Start server in a goroutine
	go func() {
		if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
			// Use provided TLS certificates
			if err := s.server.ServeTLS(listener, s.config.TLSCertFile, s.config.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("Server error: %v", err)
			}
		} else {
			// Use HTTP (for development/testing without certs)
			if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
				log.Printf("Server error: %v", err)
			}
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server
func (s *APIServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Helper function to write error responses
func writeError(w http.ResponseWriter, statusCode int, errorCode, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResponse := APIError{
		Code:       errorCode,
		Message:    message,
		RequestID:  requestID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	}

	json.NewEncoder(w).Encode(errResponse)
}

// validateToken is a placeholder for token validation
func validateToken(token string) bool {
	// TODO: Implement actual passkey verification
	// For now, accept any non-empty token
	return len(token) > 0
}

// extractUserID is a placeholder for extracting user ID from token
func extractUserID(token string) string {
	// TODO: Parse JWT or token to extract user ID
	// For now, return a placeholder
	return "user-" + token[:8]
}

func main() {
	// Load configuration
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8000"
	}

	tlsCert := os.Getenv("TLS_CERT_FILE")
	tlsKey := os.Getenv("TLS_KEY_FILE")

	allowedOrigins := []string{
		"http://localhost:3000",   // Local development
		"http://localhost:8000",   // Local API
		"https://app.example.com", // iOS app domain
	}

	config := APIConfig{
		Port:             port,
		TLSCertFile:      tlsCert,
		TLSKeyFile:       tlsKey,
		AllowedOrigins:   allowedOrigins,
		RequireAuthToken: true,
		ShutdownTimeout:  10 * time.Second,
	}

	server := NewAPIServer(config)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), server.config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
