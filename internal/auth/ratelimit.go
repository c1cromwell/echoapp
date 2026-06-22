package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimitStore is an atomic, shared (e.g. Redis) sliding-window backend. When set,
// auth limits (OTP sends, login, refresh) are enforced across all instances rather
// than per-process — closing the cross-instance bypass under horizontal scaling.
// *infra.RedisClient satisfies this structurally.
type RateLimitStore interface {
	AllowRate(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfter time.Duration, err error)
}

// AuthRateLimitConfig defines a rate limit rule.
type AuthRateLimitConfig struct {
	Name   string
	Limit  int
	Window time.Duration
}

// AuthRateLimiter implements sliding window rate limiting, in-process by default and
// distributed when a RateLimitStore is attached via SetStore.
type AuthRateLimiter struct {
	mu      sync.Mutex
	windows map[string][]time.Time // key -> timestamps of requests
	store   RateLimitStore         // optional; enforced cross-instance when set
}

// NewAuthRateLimiter creates a new rate limiter.
func NewAuthRateLimiter() *AuthRateLimiter {
	return &AuthRateLimiter{
		windows: make(map[string][]time.Time),
	}
}

// SetStore enables distributed (Redis-backed) enforcement. A store error at request
// time falls back to the in-process window rather than failing the request.
func (rl *AuthRateLimiter) SetStore(store RateLimitStore) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.store = store
}

// Check tests whether a request should be allowed under the given config.
// Returns nil if allowed, AuthError if rate limited.
func (rl *AuthRateLimiter) Check(key string, cfg AuthRateLimitConfig) *AuthError {
	// Distributed path (lock-free so the Redis round-trip doesn't serialize checks).
	rl.mu.Lock()
	store := rl.store
	rl.mu.Unlock()
	if store != nil {
		if allowed, _, retryAfter, err := store.AllowRate(context.Background(), key, cfg.Limit, cfg.Window); err == nil {
			if allowed {
				return nil
			}
			ra := int(retryAfter.Seconds())
			if ra < 1 {
				ra = 1
			}
			e := NewAuthError(ErrCodeGlobalRateLimit, 429)
			e.RetryAfter = &ra
			return e
		}
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-cfg.Window)

	// Remove expired entries
	entries := rl.windows[key]
	valid := entries[:0]
	for _, t := range entries {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	// Check limit
	if len(valid) >= cfg.Limit {
		retryAfter := int(cfg.Window.Seconds())
		if len(valid) > 0 {
			oldest := valid[0]
			retryAfter = int(oldest.Add(cfg.Window).Sub(now).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
		}
		err := NewAuthError(ErrCodeGlobalRateLimit, 429)
		err.RetryAfter = &retryAfter
		rl.windows[key] = valid
		return err
	}

	// Record request
	valid = append(valid, now)
	rl.windows[key] = valid

	return nil
}

// Remaining returns the number of requests remaining in the current window.
func (rl *AuthRateLimiter) Remaining(key string, cfg AuthRateLimitConfig) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-cfg.Window)

	entries := rl.windows[key]
	count := 0
	for _, t := range entries {
		if t.After(windowStart) {
			count++
		}
	}

	remaining := cfg.Limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// Reset clears the rate limit for a key.
func (rl *AuthRateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.windows, key)
}

// --- Pre-defined Rate Limit Configurations ---

// RateLimitOTPSend limits OTP sends per phone.
var RateLimitOTPSend = AuthRateLimitConfig{
	Name:   "otp_send",
	Limit:  5,
	Window: time.Hour,
}

// RateLimitOTPVerify limits OTP verification attempts per session.
var RateLimitOTPVerify = AuthRateLimitConfig{
	Name:   "otp_verify",
	Limit:  5,
	Window: 10 * time.Minute,
}

// RateLimitLoginIP limits login attempts per IP.
var RateLimitLoginIP = AuthRateLimitConfig{
	Name:   "login_ip",
	Limit:  10,
	Window: time.Minute,
}

// RateLimitLoginAccount limits login attempts per account.
var RateLimitLoginAccount = AuthRateLimitConfig{
	Name:   "login_account",
	Limit:  20,
	Window: time.Hour,
}

// RateLimitRefresh limits token refresh requests per DID.
var RateLimitRefresh = AuthRateLimitConfig{
	Name:   "refresh",
	Limit:  30,
	Window: time.Hour,
}

// RateLimitStepUp limits step-up authentication per DID.
var RateLimitStepUp = AuthRateLimitConfig{
	Name:   "step_up",
	Limit:  5,
	Window: time.Hour,
}

// RateLimitRecovery limits recovery initiation per phone.
var RateLimitRecovery = AuthRateLimitConfig{
	Name:   "recovery",
	Limit:  3,
	Window: 24 * time.Hour,
}

// RateLimitAPIUnverified limits API calls for unverified users.
var RateLimitAPIUnverified = AuthRateLimitConfig{
	Name:   "api_unverified",
	Limit:  100,
	Window: time.Minute,
}

// RateLimitAPIVerified limits API calls for verified users.
var RateLimitAPIVerified = AuthRateLimitConfig{
	Name:   "api_verified",
	Limit:  500,
	Window: time.Minute,
}

// FormatRateLimitKey builds a Redis-style rate limit key.
func FormatRateLimitKey(prefix string, identifier string) string {
	return fmt.Sprintf("rl:%s:%s", prefix, identifier)
}
