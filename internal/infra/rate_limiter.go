package infra

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimitStore is an atomic, shared (e.g. Redis) sliding-window backend. When set
// on a RateLimiter, limits are enforced across all instances instead of per-process.
type RateLimitStore interface {
	AllowRate(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfter time.Duration, err error)
}

// RateLimitExceededError carries the Retry-After duration so the HTTP layer
// can set the Retry-After header per RFC 6585 §4.
type RateLimitExceededError struct {
	RetryAfter time.Duration
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded; retry after %.0fs", e.RetryAfter.Seconds())
}

// ErrRateLimitExceeded is the sentinel value for callers that only care about
// the category (not the retry delay).
var ErrRateLimitExceeded = &RateLimitExceededError{RetryAfter: 60 * time.Second}

// RateLimitConfig defines the rate limit for a specific action type.
type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

// SubscriptionTier controls which rate limit table is applied.
type SubscriptionTier string

const (
	TierBase SubscriptionTier = "base"
	TierVIP  SubscriptionTier = "vip"
)

// DefaultRateLimits returns the per-DID base rate limits from the architecture spec.
func DefaultRateLimits() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		"api_request":     {MaxRequests: 100, Window: time.Minute},
		"message_send":    {MaxRequests: 60, Window: time.Minute},
		"websocket_msg":   {MaxRequests: 60, Window: time.Minute},
		"sync_push":       {MaxRequests: 120, Window: time.Minute},
		"backup_push":     {MaxRequests: 20, Window: time.Hour},
		"group_mutation":  {MaxRequests: 30, Window: time.Minute},
		"reward_claim":    {MaxRequests: 10, Window: 24 * time.Hour},
		"data_submission": {MaxRequests: 10, Window: time.Minute},
	}
}

// VIPRateLimits returns the 2× limits for VIP subscribers ($9.99/mo).
func VIPRateLimits() map[string]RateLimitConfig {
	base := DefaultRateLimits()
	vip := make(map[string]RateLimitConfig, len(base))
	for action, cfg := range base {
		vip[action] = RateLimitConfig{MaxRequests: cfg.MaxRequests * 2, Window: cfg.Window}
	}
	return vip
}

// RateLimiter provides per-DID rate limiting using a sliding window approach.
// Supports base and VIP subscription tiers (WO-44).
type RateLimiter struct {
	mu         sync.Mutex
	baseLimits map[string]RateLimitConfig
	vipLimits  map[string]RateLimitConfig
	limits     map[string]RateLimitConfig // kept for backward compat
	buckets    map[string]*tokenBucket
	store      RateLimitStore // optional; when set, enforced cross-instance
}

// SetStore enables distributed (e.g. Redis-backed) enforcement. When the store
// errors at request time the limiter falls back to its in-process window, so a
// Redis blip degrades to per-instance limiting rather than failing requests.
func (rl *RateLimiter) SetStore(store RateLimitStore) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.store = store
}

type tokenBucket struct {
	tokens    int
	maxTokens int
	window    time.Duration
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter with the given base limit configurations.
// VIP limits default to 2× base.
func NewRateLimiter(limits map[string]RateLimitConfig) *RateLimiter {
	vip := make(map[string]RateLimitConfig, len(limits))
	for k, v := range limits {
		vip[k] = RateLimitConfig{MaxRequests: v.MaxRequests * 2, Window: v.Window}
	}
	return &RateLimiter{
		baseLimits: limits,
		vipLimits:  vip,
		limits:     limits,
		buckets:    make(map[string]*tokenBucket),
	}
}

// Check verifies that the given DID has not exceeded the base rate limit for the action.
// Returns nil if allowed, *RateLimitExceededError if over limit.
func (rl *RateLimiter) Check(did, action string) error {
	return rl.CheckTiered(did, action, TierBase)
}

// CheckTiered is like Check but applies VIP limits when tier == TierVIP.
func (rl *RateLimiter) CheckTiered(did, action string, tier SubscriptionTier) error {
	// baseLimits/vipLimits are set once at construction and never mutated, so they
	// are safe to read without the lock.
	limits := rl.baseLimits
	if tier == TierVIP {
		limits = rl.vipLimits
	}

	limit, ok := limits[action]
	if !ok {
		return nil // Unknown action — allow
	}

	key := string(tier) + ":" + did + ":" + action

	// Distributed path: enforce across instances via the shared store, lock-free so a
	// Redis round-trip doesn't serialize all checks. On error, fall back to the
	// in-process window (degraded but still protected).
	rl.mu.Lock()
	store := rl.store
	rl.mu.Unlock()
	if store != nil {
		if allowed, _, retryAfter, err := store.AllowRate(context.Background(), key, limit.MaxRequests, limit.Window); err == nil {
			if allowed {
				return nil
			}
			if retryAfter <= 0 {
				retryAfter = limit.Window
			}
			return &RateLimitExceededError{RetryAfter: retryAfter}
		}
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()
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

	if time.Since(bucket.lastReset) >= bucket.window {
		bucket.tokens = bucket.maxTokens - 1
		bucket.lastReset = time.Now()
		return nil
	}

	if bucket.tokens <= 0 {
		retryAfter := bucket.window - time.Since(bucket.lastReset)
		return &RateLimitExceededError{RetryAfter: retryAfter}
	}

	bucket.tokens--
	return nil
}

// Remaining returns the number of requests remaining for a DID/action pair
// at the base subscription tier.
func (rl *RateLimiter) Remaining(did, action string) int {
	return rl.RemainingTiered(did, action, TierBase)
}

// RemainingTiered returns the remaining requests for the given tier.
func (rl *RateLimiter) RemainingTiered(did, action string, tier SubscriptionTier) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limits := rl.baseLimits
	if tier == TierVIP {
		limits = rl.vipLimits
	}

	key := string(tier) + ":" + did + ":" + action
	bucket, exists := rl.buckets[key]
	if !exists {
		if limit, ok := limits[action]; ok {
			return limit.MaxRequests
		}
		return -1
	}

	if time.Since(bucket.lastReset) >= bucket.window {
		return bucket.maxTokens
	}

	return bucket.tokens
}
