// Package api provides the shared HTTP router, middleware, and handlers
// used by both the production server and integration tests.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thechadcromwell/echoapp/internal/auth"
	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
	"github.com/thechadcromwell/echoapp/internal/services/onboarding"
	"github.com/thechadcromwell/echoapp/pkg/credentials"
	"github.com/thechadcromwell/echoapp/pkg/credentials/oidc4vc"
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
	AllowedOrigins       []string
	StartTime            time.Time
	TokenValidator       func(token string) bool
	UserIDExtractor      func(token string) string
	WSHub                *Hub          // WebSocket hub for real-time messaging
	V3                   *V3Handlers   // V3 API handlers (blueprint services)
	DIDRegistry          DIDRegistry   // did:key binding store (WO-230 / WO-278)
	CredentialStatusPool *pgxpool.Pool // WO-274 durable VC status list slots (optional)
	Redis                *infra.RedisClient
	RateLimiter          *infra.RateLimiter         // WO-44 per-DID tiered rate limiting (optional)
	PublicRateLimiter    *infra.RateLimiter         // S5: per-IP throttle for pre-auth public endpoints (optional)
	SMSProvider          infra.SMSProvider          // Wave 12 SMS OTP recovery (optional; stub when nil)
	IdentityL1           *metagraph.MetagraphClient // WO-274 trust-tier commitments
	DataL1               *metagraph.MetagraphClient // WO-230 Data L1 Merkle proxy (optional)
	CredentialService    *credentials.Service       // WO-274 VC issuance (optional)
	OIDC                 *gin.Engine                // OpenID4VCI issuer + verifier mount (optional)
	OIDCVerifier         *oidc4vc.Verifier          // OIDC4VC presentation verifier (optional)
	OIDCVerifierBaseURL  string                     // Public verifier base URL for iOS wallet handoff
	TrustRegistry        *onboarding.TrustRegistryService // WO-118 issuer trust registry
	tokenService         *auth.TokenService         // ES256 JWT token service

	enrollmentVCMu       sync.Mutex
	enrollmentVCSessions map[string]enrollmentVCSession

	// smsSessions is an in-memory fallback OTP-session store used only when
	// Redis is not configured (dev/test); prod uses Redis. Guarded by smsSessionsMu.
	smsSessions   map[string]smsOTPSession
	smsSessionsMu sync.Mutex
}

// NewRouter creates a Router with production-grade ES256 JWT validation.
// It also creates and starts a WebSocket hub.
func NewRouter(allowedOrigins []string) *Router {
	hub := NewHub()
	go hub.Run()

	tokenService, err := auth.NewTokenService()
	if err != nil {
		log.Fatalf("Failed to create token service: %v", err)
	}

	rt := &Router{
		AllowedOrigins: allowedOrigins,
		StartTime:      time.Now(),
		WSHub:          hub,
		DIDRegistry:    NewMemoryDIDRegistry(),
		TrustRegistry:  onboarding.NewTrustRegistryService(),
		tokenService:   tokenService,
	}

	// Wire up JWT-based validation
	rt.TokenValidator = func(token string) bool {
		_, err := tokenService.ValidateAccessToken(token)
		return err == nil
	}
	rt.UserIDExtractor = func(token string) string {
		claims, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			return ""
		}
		return claims.Subject
	}

	return rt
}

// TokenService returns the router's JWT token service for issuing tokens.
func (rt *Router) TokenService() *auth.TokenService {
	return rt.tokenService
}

func isOpenIDCredentialPath(path string) bool {
	if isOpenIDCredentialIssuerPath(path) {
		return true
	}
	switch {
	case strings.HasPrefix(path, "/.well-known/openid-credential-verifier"):
		return true
	case strings.HasPrefix(path, "/verification/"):
		return true
	case strings.HasPrefix(path, "/presentation_definition/"):
		return true
	default:
		return false
	}
}

func isOpenIDCredentialIssuerPath(path string) bool {
	switch {
	case path == "/notification":
		return true
	case strings.HasPrefix(path, "/.well-known/openid-credential-issuer"):
		return true
	case strings.HasPrefix(path, "/.well-known/oauth-authorization-server"):
		return true
	case strings.HasPrefix(path, "/oauth/"):
		return true
	case path == "/credential" || strings.HasPrefix(path, "/credential/"):
		return true
	default:
		return false
	}
}

// Handler returns the fully wrapped http.Handler with all middleware applied.
func (rt *Router) Handler() http.Handler {
	core := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rt.OIDC != nil && isOpenIDCredentialPath(r.URL.Path) {
			rt.OIDC.ServeHTTP(w, r)
			return
		}
		switch {
		case r.URL.Path == "/health":
			rt.handleHealth(w, r)
		case r.URL.Path == "/ws":
			// WebSocket upgrade — handled before auth middleware wraps the response
			ServeWS(rt.WSHub, rt.UserIDExtractor, w, r)
		case r.URL.Path == "/identity/register":
			rt.handleIdentityRegister(w, r)
		case r.URL.Path == "/identity/devices":
			rt.handleIdentityAddDevice(w, r)
		case r.URL.Path == "/identity/devices/token":
			rt.handleIdentityDeviceToken(w, r)
		case strings.HasPrefix(r.URL.Path, "/identity/devices/"):
			did := strings.TrimPrefix(r.URL.Path, "/identity/devices/")
			rt.handleIdentityListDevices(w, r, did)
		case r.URL.Path == "/identity/trust-tier/commitment":
			rt.handleTrustTierCommitment(w, r)
		case strings.HasPrefix(r.URL.Path, "/identity/resolve/"):
			did := strings.TrimPrefix(r.URL.Path, "/identity/resolve/")
			rt.handleIdentityResolve(w, r, did)
		case strings.HasPrefix(r.URL.Path, "/identity/credentials/status/"):
			rt.handleCredentialVCStatus(w, r)
		case r.URL.Path == "/identity/credentials":
			rt.handleIdentityCredentials(w, r)
		case strings.HasPrefix(r.URL.Path, "/identity/"):
			did := strings.TrimPrefix(r.URL.Path, "/identity/")
			rt.handleIdentityResolve(w, r, did)
		case r.URL.Path == "/v1/crypto/server-key":
			rt.handleServerKey(w, r)
		case r.URL.Path == "/v1/auth/sms-recovery/register":
			rt.handleSMSRecoveryRegister(w, r)
		case r.URL.Path == "/v1/auth/sms-recovery/verify":
			rt.handleSMSRecoveryVerify(w, r)
		case r.URL.Path == "/v1/auth/vip-verify":
			rt.handleVIPVerify(w, r)
		case r.URL.Path == "/v1/auth/login/challenge":
			rt.handleLoginChallenge(w, r)
		case r.URL.Path == "/v1/auth/sms-recovery/challenge":
			rt.handleSMSRecoveryChallenge(w, r)
		case strings.HasPrefix(r.URL.Path, "/v1/"):
			rt.handleV1(w, r)
		case strings.HasPrefix(r.URL.Path, "/v2/"):
			rt.handleV2(w, r)
		case r.URL.Path == "/v3/auth/refresh":
			rt.handleAuthRefresh(w, r)
		case r.URL.Path == "/v3/auth/revoke":
			rt.handleAuthRevoke(w, r)
		case strings.HasPrefix(r.URL.Path, "/v3/"):
			rt.handleV3(w, r)
		default:
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint not found", r.Header.Get("X-Request-ID"))
		}
	})

	// Middleware chain: CORS -> Auth -> RateLimit -> RequestID -> core
	return rt.corsMiddleware(rt.authMiddleware(rt.rateLimitMiddleware(rt.requestIDMiddleware(core))))
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

// publicPaths are exempt from Bearer token authentication.
// These are registration and restore endpoints that new devices call before they have a token.
var publicPaths = map[string]bool{
	"/v1/auth/register-did":            true,
	"/v1/auth/restore-challenge":       true,
	"/v1/auth/restore-did":             true,
	"/v1/enrollment/vc/start":          true,
	"/v1/enrollment/mdl/start":         true,
	"/v1/enrollment/idv/start":         true,
	"/v1/enrollment/idv/await":         true,
	"/v1/data-l1/merkle-roots":         true,
	"/v1/phase1/trust-tier-commitment": true,
	"/v1/crypto/server-key":            true, // WO-13: public key endpoint, no auth
	"/v1/users/check-username":         true, // WO-14: pre-auth username availability check
	"/v1/onboarding/session/start":     true, // WO-203: universal onboarding orchestrator (Wave 0.6)
	"/v3/auth/refresh":                 true, // A.2: refresh token is the credential; no access token yet
	"/v1/auth/sms-recovery/register":   true, // Wave 12: phone commitment registration
	"/v1/auth/sms-recovery/verify":     true, // Wave 12: OTP verification
	"/v1/auth/sms-recovery/challenge":  true, // Wave 12: recovery challenge
	"/v1/auth/vip-verify":              true, // Phase 1: VIP identity verification result
	"/v1/auth/login/challenge":         true, // Phase 1: nonce for passkey signing (public)
	"/identity/register":               true,
	"/identity/devices":                true,
}

// publicPreAuthAction is the rate-limit action key for unauthenticated endpoints.
const publicPreAuthAction = "public_pre_auth"

// ipThrottledPublicPaths are pre-auth endpoints throttled per client IP (S5) to
// curb username enumeration, OTP abuse, and registration spam. They have no DID
// to key the per-DID limiter on, so the IP is used instead.
var ipThrottledPublicPaths = map[string]bool{
	"/v1/users/check-username":        true,
	"/v1/onboarding/session/start":    true,
	"/v1/auth/login/challenge":        true,
	"/identity/register":              true,
	"/v1/auth/sms-recovery/register":  true,
	"/v1/auth/sms-recovery/verify":    true,
	"/v1/auth/sms-recovery/challenge": true,
	"/v3/auth/refresh":                true,
}

// clientIP extracts the caller's IP for rate limiting. It honors the leftmost
// X-Forwarded-For entry (set by a trusted reverse proxy) and falls back to the
// connection's remote address.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// identityRequestExemptFromAuth lists unauthenticated Identity routes (did:key resolution, etc.).
// Authenticated Identity routes such as POST /identity/trust-tier/commitment are not exempt.
func identityRequestExemptFromAuth(path, method string) bool {
	switch {
	case path == "/identity/register" || path == "/identity/devices":
		return true
	case strings.HasPrefix(path, "/identity/resolve/"):
		return true
	case method == http.MethodGet && strings.HasPrefix(path, "/identity/credentials/status/"):
		return true
	case method == http.MethodGet && path != "/identity/credentials" && strings.HasPrefix(path, "/identity/"):
		return true
	default:
		return false
	}
}

func trustRegistryPublicRead(path, method string) bool {
	if method != http.MethodGet {
		return false
	}
	return path == "/v1/trust-registry/issuers" || strings.HasPrefix(path, "/v1/trust-registry/issuers/")
}

func trustRegistryAdminMutation(path, method string, r *http.Request) bool {
	if method != http.MethodPost {
		return false
	}
	return path == "/v1/trust-registry/issuers" || strings.Contains(path, "/v1/trust-registry/issuers/")
}

func onboardingSessionPublic(path, method string) bool {
	if !strings.HasPrefix(path, "/v1/onboarding/session") {
		return false
	}
	switch method {
	case http.MethodPost:
		return path == "/v1/onboarding/session/start" || strings.HasSuffix(path, "/advance")
	case http.MethodGet:
		return strings.HasPrefix(path, "/v1/onboarding/session/")
	default:
		return false
	}
}

func (rt *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health check, WebSocket, public registration endpoints, and
		// did:key resolution (which is local-only and inherently public) bypass auth.
		if r.URL.Path == "/health" || r.URL.Path == "/ws" || publicPaths[r.URL.Path] ||
			identityRequestExemptFromAuth(r.URL.Path, r.Method) ||
			isOpenIDCredentialPath(r.URL.Path) ||
			trustRegistryPublicRead(r.URL.Path, r.Method) ||
			trustRegistryAdminMutation(r.URL.Path, r.Method, r) ||
			onboardingSessionPublic(r.URL.Path, r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		reqID := r.Header.Get("X-Request-ID")

		// --- Passkey auth (WO-1): X-Sender-DID + X-Signature over SHA-256(body) ---
		// Primary auth mechanism. Falls through to JWT only when X-Sender-DID is absent,
		// allowing backward-compat with service tokens during the Phase 1 transition.
		if senderDID := r.Header.Get(headerSenderDID); senderDID != "" {
			keys, err := rt.resolveDeviceKeys(r.Context(), senderDID)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "AUTH_UNKNOWN_DID", "no device keys found for DID", reqID)
				return
			}
			if err := rt.verifyPasskeyAuth(r, keys); err != nil {
				if pke, ok := err.(*passkeyAuthError); ok {
					WriteError(w, http.StatusUnauthorized, pke.code, pke.msg, reqID)
				} else {
					WriteError(w, http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", err.Error(), reqID)
				}
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyUserID, senderDID)
			ctx = context.WithValue(ctx, ContextKeyRequestID, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// --- JWT auth (backward compat / server tokens) ---
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteError(w, http.StatusUnauthorized, "MISSING_AUTH", "Authorization header or X-Sender-DID required", reqID)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteError(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization must be Bearer token", reqID)
			return
		}

		token := parts[1]
		if !rt.TokenValidator(token) {
			WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", reqID)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, rt.UserIDExtractor(token))
		ctx = context.WithValue(ctx, ContextKeyRequestID, reqID)
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Sender-DID, X-Signature, X-Timestamp, X-Identity-Signature")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Retry-After, X-RateLimit-Remaining")
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

// rateLimitMiddleware enforces WO-44 per-DID tiered rate limits.
// It runs after auth so the DID is available in context.
// Health, WebSocket and public registration paths are exempt.
func (rt *Router) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		reqID := r.Header.Get("X-Request-ID")

		// S5: per-IP throttle for pre-auth public endpoints (independent of the
		// per-DID limiter, which has no DID to key on for these routes).
		if rt.PublicRateLimiter != nil && ipThrottledPublicPaths[path] {
			ip := clientIP(r)
			if err := rt.PublicRateLimiter.Check(ip, publicPreAuthAction); err != nil {
				if rle, ok := err.(*infra.RateLimitExceededError); ok {
					w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rle.RetryAfter.Seconds()))
				}
				WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", reqID)
				return
			}
		}

		if rt.RateLimiter == nil {
			next.ServeHTTP(w, r)
			return
		}

		if path == "/health" || path == "/ws" || publicPaths[path] ||
			identityRequestExemptFromAuth(path, r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		did, _ := r.Context().Value(ContextKeyUserID).(string)
		if did == "" {
			// Unauthenticated request — auth middleware will reject it; skip rate limiting.
			next.ServeHTTP(w, r)
			return
		}

		remaining := rt.RateLimiter.Remaining(did, "api_request")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if err := rt.RateLimiter.Check(did, "api_request"); err != nil {
			if rle, ok := err.(*infra.RateLimitExceededError); ok {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rle.RetryAfter.Seconds()))
			}
			WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", reqID)
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
	if strings.HasPrefix(r.URL.Path, "/v1/trust-registry/") {
		rt.handleTrustRegistry(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v1/onboarding/") {
		rt.handleOnboardingSession(w, r)
		return
	}
	switch r.URL.Path {
	case "/v1/users":
		rt.v1GetUsers(w, r)
	case "/v1/users/check-username":
		rt.handleCheckUsername(w, r)
	case "/v1/users/profile":
		rt.v1GetProfile(w, r)

	// --- Auth endpoints (REQ-INFRA-004) ---
	case "/v1/auth/step-up":
		rt.handleStepUp(w, r)
	case "/v1/auth/register-did":
		rt.handleRegisterDID(w, r)
	case "/v1/auth/restore-challenge":
		rt.handleRestoreChallenge(w, r)
	case "/v1/auth/restore-did":
		rt.handleRestoreDID(w, r)

	// --- Enrollment tail endpoints (REQ-INFRA-004) ---
	case "/v1/enrollment/passkey":
		rt.handleRegisterPasskey(w, r)
	case "/v1/enrollment/vc/start", "/v1/enrollment/vc/finish":
		rt.handleEnrollmentVC(w, r)
	case "/v1/enrollment/mdl/start", "/v1/enrollment/mdl/finish":
		rt.handleEnrollmentMDL(w, r)
	case "/v1/enrollment/idv/start", "/v1/enrollment/idv/await":
		rt.handleEnrollmentIDV(w, r)
	case "/v1/data-l1/merkle-roots":
		rt.handleDataL1MerkleRoots(w, r)
	case "/v1/phase1/trust-tier-commitment":
		rt.handlePhase1TrustTierCommitment(w, r)

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
