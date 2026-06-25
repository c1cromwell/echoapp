package bots

import (
	"sync"
	"time"
)

// RateLimiter enforces per-bot message velocity (WO-11: 100/min).
type RateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	limit    int
	counters map[string][]time.Time
}

// NewRateLimiter creates a bot rate limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 100
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		window:   window,
		limit:    limit,
		counters: make(map[string][]time.Time),
	}
}

// Allow reports whether botDID may send another message now.
func (r *RateLimiter) Allow(botDID string) bool {
	if r == nil || botDID == "" {
		return false
	}
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := now.Add(-r.window)
	times := r.counters[botDID]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= r.limit {
		r.counters[botDID] = kept
		return false
	}
	r.counters[botDID] = append(kept, now)
	return true
}
