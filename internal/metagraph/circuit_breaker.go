package metagraph

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the breaker is open and a call is short-circuited.
// Callers should treat it as "metagraph temporarily unavailable" and not retry tightly.
var ErrCircuitOpen = errors.New("metagraph: circuit breaker open")

type breakerState int

const (
	breakerClosed breakerState = iota
	breakerOpen
	breakerHalfOpen
)

// CircuitBreaker fast-fails calls to a failing dependency so callers don't each wait
// for the full request timeout (which would pile up goroutines and stall the relay/
// anchoring path). After `threshold` consecutive failures it opens for `cooldown`,
// then admits a single probe (half-open); a probe success closes it, a failure
// re-opens it.
type CircuitBreaker struct {
	mu        sync.Mutex
	state     breakerState
	failures  int
	threshold int
	cooldown  time.Duration
	openedAt  time.Time
}

// NewCircuitBreaker creates a breaker that opens after `threshold` consecutive
// failures and stays open for `cooldown`.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	if threshold < 1 {
		threshold = 1
	}
	return &CircuitBreaker{state: breakerClosed, threshold: threshold, cooldown: cooldown}
}

// Allow reports whether a call may proceed. When open it admits exactly one probe
// after the cooldown elapses (transitioning to half-open) and rejects the rest.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case breakerOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = breakerHalfOpen // admit one probe
			return true
		}
		return false
	case breakerHalfOpen:
		return false // a probe is already in flight
	default:
		return true
	}
}

// Success records a successful call and closes the breaker.
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = breakerClosed
}

// Failure records a failed call, opening the breaker once the threshold is reached
// (or immediately if a half-open probe fails).
func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == breakerHalfOpen || cb.failures >= cb.threshold {
		cb.state = breakerOpen
		cb.openedAt = time.Now()
	}
}

// State returns "closed", "open", or "half-open" for observability.
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case breakerOpen:
		return "open"
	case breakerHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}

// guarded runs fn under the breaker: it short-circuits with ErrCircuitOpen when open,
// and records the outcome to drive state transitions.
func (c *MetagraphClient) guarded(fn func() (string, error)) (string, error) {
	if c.breaker != nil && !c.breaker.Allow() {
		return "", ErrCircuitOpen
	}
	res, err := fn()
	if c.breaker != nil {
		if err != nil {
			c.breaker.Failure()
		} else {
			c.breaker.Success()
		}
	}
	return res, err
}
