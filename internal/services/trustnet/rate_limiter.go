package trustnet

import (
	"sync"
	"time"
)

// TrustOperation represents a type of trust-related operation
type TrustOperation string

const (
	OpEndorsement   TrustOperation = "endorsement"
	OpPromotion     TrustOperation = "promotion"
	OpReport        TrustOperation = "report"
	OpVerification  TrustOperation = "verification"
	OpDispute       TrustOperation = "dispute"
)

// TrustOpLimits defines the rate limits for trust operations
var TrustOpLimits = map[TrustOperation]OpLimit{
	OpEndorsement:  {Daily: 5, Cooldown: 0},
	OpPromotion:    {Daily: 10, Cooldown: 0},
	OpReport:       {Daily: 10, Cooldown: 0},           // plus per-target: 1 ever
	OpVerification: {Daily: 3, Cooldown: 24 * time.Hour}, // 24h after failure
	OpDispute:      {Daily: 1, Cooldown: 90 * 24 * time.Hour},
}

// OpLimit defines limits for a single operation type
type OpLimit struct {
	Daily    int
	Cooldown time.Duration
}

// TrustOpUsage tracks usage of a single operation type
type TrustOpUsage struct {
	Date       string
	Count      int
	LastUsedAt time.Time
	// For per-target tracking (e.g., reports)
	Targets map[string]bool
}

// TrustRateLimiter manages rate limits for trust operations
type TrustRateLimiter struct {
	mu    sync.Mutex
	usage map[string]map[TrustOperation]*TrustOpUsage // userDID -> operation -> usage
}

// NewTrustRateLimiter creates a new trust operation rate limiter
func NewTrustRateLimiter() *TrustRateLimiter {
	return &TrustRateLimiter{
		usage: make(map[string]map[TrustOperation]*TrustOpUsage),
	}
}

// Check checks if a user can perform an operation
func (r *TrustRateLimiter) Check(userDID string, op TrustOperation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit, ok := TrustOpLimits[op]
	if !ok {
		return nil // unknown operation, allow
	}

	usage := r.getOrCreateUsage(userDID, op)
	today := time.Now().Format("2006-01-02")

	// Reset if new day
	if usage.Date != today {
		usage.Date = today
		usage.Count = 0
	}

	if usage.Count >= limit.Daily {
		return ErrTrustOpRateLimited
	}

	// Check cooldown
	if limit.Cooldown > 0 && !usage.LastUsedAt.IsZero() {
		if time.Since(usage.LastUsedAt) < limit.Cooldown {
			return ErrTrustOpRateLimited
		}
	}

	return nil
}

// CheckWithTarget checks rate limit including per-target limits (for reports)
func (r *TrustRateLimiter) CheckWithTarget(userDID string, op TrustOperation, targetDID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit, ok := TrustOpLimits[op]
	if !ok {
		return nil
	}

	usage := r.getOrCreateUsage(userDID, op)
	today := time.Now().Format("2006-01-02")

	if usage.Date != today {
		usage.Date = today
		usage.Count = 0
	}

	if usage.Count >= limit.Daily {
		return ErrTrustOpRateLimited
	}

	// Per-target check (e.g., can only report each user once ever)
	if op == OpReport && usage.Targets[targetDID] {
		return ErrTrustOpRateLimited
	}

	return nil
}

// Record records that an operation was performed
func (r *TrustRateLimiter) Record(userDID string, op TrustOperation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	usage := r.getOrCreateUsage(userDID, op)
	today := time.Now().Format("2006-01-02")

	if usage.Date != today {
		usage.Date = today
		usage.Count = 0
	}

	usage.Count++
	usage.LastUsedAt = time.Now()
}

// RecordWithTarget records an operation with a specific target
func (r *TrustRateLimiter) RecordWithTarget(userDID string, op TrustOperation, targetDID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	usage := r.getOrCreateUsage(userDID, op)
	today := time.Now().Format("2006-01-02")

	if usage.Date != today {
		usage.Date = today
		usage.Count = 0
	}

	usage.Count++
	usage.LastUsedAt = time.Now()
	usage.Targets[targetDID] = true
}

// GetRemaining returns the number of remaining operations for today
func (r *TrustRateLimiter) GetRemaining(userDID string, op TrustOperation) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit, ok := TrustOpLimits[op]
	if !ok {
		return -1
	}

	usage := r.getOrCreateUsage(userDID, op)
	today := time.Now().Format("2006-01-02")

	if usage.Date != today {
		return limit.Daily
	}

	remaining := limit.Daily - usage.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetCooldownRemaining returns the remaining cooldown time for an operation
func (r *TrustRateLimiter) GetCooldownRemaining(userDID string, op TrustOperation) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit, ok := TrustOpLimits[op]
	if !ok || limit.Cooldown == 0 {
		return 0
	}

	usage := r.getOrCreateUsage(userDID, op)
	if usage.LastUsedAt.IsZero() {
		return 0
	}

	elapsed := time.Since(usage.LastUsedAt)
	if elapsed >= limit.Cooldown {
		return 0
	}
	return limit.Cooldown - elapsed
}

// HasTargeted checks if a user has already targeted a specific DID with an operation
func (r *TrustRateLimiter) HasTargeted(userDID string, op TrustOperation, targetDID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	usage := r.getOrCreateUsage(userDID, op)
	return usage.Targets[targetDID]
}

func (r *TrustRateLimiter) getOrCreateUsage(userDID string, op TrustOperation) *TrustOpUsage {
	if r.usage[userDID] == nil {
		r.usage[userDID] = make(map[TrustOperation]*TrustOpUsage)
	}
	if r.usage[userDID][op] == nil {
		r.usage[userDID][op] = &TrustOpUsage{
			Targets: make(map[string]bool),
		}
	}
	return r.usage[userDID][op]
}
