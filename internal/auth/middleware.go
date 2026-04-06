package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const (
	claimsKey     contextKey = "auth_claims"
	deviceInfoKey contextKey = "device_info"
)

// GetClaimsFromContext retrieves token claims from the request context.
func GetClaimsFromContext(ctx context.Context) *TokenClaims {
	claims, _ := ctx.Value(claimsKey).(*TokenClaims)
	return claims
}

// GetDeviceInfoFromContext retrieves device info from the request context.
func GetDeviceInfoFromContext(ctx context.Context) *DeviceInfo {
	info, _ := ctx.Value(deviceInfoKey).(*DeviceInfo)
	return info
}

// AuthMiddleware provides HTTP middleware for JWT validation and device binding.
type AuthMiddleware struct {
	tokenService *TokenService
	rateLimiter  *AuthRateLimiter
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(tokenService *TokenService, rateLimiter *AuthRateLimiter) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		rateLimiter:  rateLimiter,
	}
}

// ExtractDeviceInfo parses the X-Device-Info header into context.
func ExtractDeviceInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("X-Device-Info")
		if header == "" {
			writeError(w, 400, ErrCodeDeviceIntegrity, "Missing X-Device-Info header")
			return
		}

		var info DeviceInfo
		if err := json.Unmarshal([]byte(header), &info); err != nil {
			writeError(w, 400, ErrCodeDeviceIntegrity, "Invalid X-Device-Info header")
			return
		}

		if authErr := ValidateDeviceInfo(info); authErr != nil {
			writeError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		ctx := context.WithValue(r.Context(), deviceInfoKey, &info)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateDeviceIntegrityMiddleware rejects jailbroken devices.
func ValidateDeviceIntegrityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := GetDeviceInfoFromContext(r.Context())
		if info == nil {
			writeError(w, 400, ErrCodeDeviceIntegrity, "Device info required")
			return
		}

		if authErr := ValidateDeviceIntegrity(*info); authErr != nil {
			writeError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ValidateJWT verifies the access token and adds claims to context.
func (m *AuthMiddleware) ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractBearerToken(r)
		if tokenString == "" {
			writeError(w, 401, ErrCodeTokenExpired, "Authentication required")
			return
		}

		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				writeError(w, 401, ErrCodeTokenExpired, ErrCodeTokenExpired.UserFacingMessage())
			} else {
				writeError(w, 401, ErrCodePasskeyFailed, ErrCodePasskeyFailed.UserFacingMessage())
			}
			return
		}

		// Check blocklist
		if m.tokenService.IsBlocklisted(claims.TokenID) {
			writeError(w, 401, ErrCodeRefreshInvalid, ErrCodeRefreshInvalid.UserFacingMessage())
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// BindDevice verifies the token's device_id matches the request device.
func BindDevice(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromContext(r.Context())
		info := GetDeviceInfoFromContext(r.Context())

		if claims == nil || info == nil {
			writeError(w, 401, ErrCodePasskeyFailed, "Authentication required")
			return
		}

		deviceHash := ComputeDeviceHash(*info)
		if claims.DeviceID != deviceHash {
			writeError(w, 403, ErrCodeUnknownDevice, ErrCodeUnknownDevice.UserFacingMessage())
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireStepUp checks that the token has elevated privileges for a specific action.
func RequireStepUp(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromContext(r.Context())
			if claims == nil || !claims.Elevated {
				writeError(w, 403, ErrCodeStepUpRequired, ErrCodeStepUpRequired.UserFacingMessage())
				return
			}
			if claims.ElevatedAction != action {
				writeError(w, 403, ErrCodeStepUpRequired, "Elevated token not valid for this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware applies rate limiting based on trust tier.
func (m *AuthMiddleware) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromContext(r.Context())
		if claims == nil {
			next.ServeHTTP(w, r)
			return
		}

		cfg := RateLimitAPIUnverified
		if claims.TrustTier >= 2 {
			cfg = RateLimitAPIVerified
		}

		key := FormatRateLimitKey("api", claims.Subject)
		if err := m.rateLimiter.Check(key, cfg); err != nil {
			writeError(w, 429, ErrCodeGlobalRateLimit, ErrCodeGlobalRateLimit.UserFacingMessage())
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Helpers ---

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

func writeError(w http.ResponseWriter, httpStatus int, code AuthErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	resp := ErrorResponse{
		Error: &AuthError{
			Code:    code,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}
