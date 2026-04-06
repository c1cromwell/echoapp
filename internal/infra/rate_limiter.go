package infra

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// RateLimitConfig defines the rate limit for a specific action type.
type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

// DefaultRateLimits returns the per-DID rate limits from the architecture spec.
func DefaultRateLimits() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		"api_request":     {MaxRequests: 100, Window: time.Minute},        // API Gateway
		"message_send":    {MaxRequests: 60, Window: time.Minute},         // Message relay
		"reward_claim":    {MaxRequests: 10, Window: 24 * time.Hour},      // Reward claims (daily)
		"data_submission": {MaxRequests: 10, Window: time.Minute},         // Data L1 submission
		"websocket_msg":   {MaxRequests: 60, Window: time.Minute},         // WebSocket per-connection
	}
}

// RateLimiter provides per-DID rate limiting using a sliding window approach.
type RateLimiter struct {
	mu      sync.Mutex
	limits  map[string]RateLimitConfig
	buckets map[string]*tokenBucket
}

type tokenBucket struct {
	tokens    int
	maxTokens int
	window    time.Duration
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter with the given limit configurations.
func NewRateLimiter(limits map[string]RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		limits:  limits,
		buckets: make(map[string]*tokenBucket),
	}
}

// Check verifies that the given DID has not exceeded the rate limit for the action.
// Returns nil if allowed, ErrRateLimitExceeded if over limit.
func (rl *RateLimiter) Check(did, action string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, ok := rl.limits[action]
	if !ok {
		return nil // Unknown action — allow
	}

	key := did + ":" + action
	bucket, exists := rl.buckets[key]

	if !exists {
		rl.buckets[key] = &tokenBucket{
			tokens:    limit.MaxRequests - 1,
			maxTokens: limit.MaxRequests,
			window:    limit.Window,
			lastReset: time.Now(),
		}
		return nil
	}

	// Reset bucket if window has elapsed
	if time.Since(bucket.lastReset) >= bucket.window {
		bucket.tokens = bucket.maxTokens - 1
		bucket.lastReset = time.Now()
		return nil
	}

	if bucket.tokens <= 0 {
		return ErrRateLimitExceeded
	}

	bucket.tokens--
	return nil
}

// Remaining returns the number of requests remaining for a DID/action pair.
func (rl *RateLimiter) Remaining(did, action string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := did + ":" + action
	bucket, exists := rl.buckets[key]
	if !exists {
		if limit, ok := rl.limits[action]; ok {
			return limit.MaxRequests
		}
		return -1
	}

	if time.Since(bucket.lastReset) >= bucket.window {
		return bucket.maxTokens
	}

	return bucket.tokens
}
