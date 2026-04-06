// Package infra provides infrastructure utilities shared across services.
package infra

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the current state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation — requests pass through
	CircuitOpen                         // Blocking requests — downstream is failing
	CircuitHalfOpen                     // Testing recovery — allowing one probe request
)

// String returns a human-readable state name.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

var (
	ErrCircuitOpen = errors.New("circuit breaker is open: downstream unavailable")
)

// Circuit tracks failure state for a single downstream dependency.
type Circuit struct {
	Name             string
	State            CircuitState
	FailureCount     int
	SuccessCount     int           // Consecutive successes in half-open state
	FailureThreshold int           // Failures before opening circuit
	SuccessThreshold int           // Successes in half-open before closing
	ResetTimeout     time.Duration // How long to wait before half-open probe
	LastFailure      time.Time
	LastStateChange  time.Time
}

// CircuitBreaker manages independent circuit breakers per downstream chain.
type CircuitBreaker struct {
	mu       sync.RWMutex
	circuits map[string]*Circuit
}

// NewCircuitBreaker creates breakers for each chain with default settings.
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		circuits: map[string]*Circuit{
			"data_l1": {
				Name:             "Metagraph Data L1",
				State:            CircuitClosed,
				FailureThreshold: 5,
				SuccessThreshold: 2,
				ResetTimeout:     30 * time.Second,
				LastStateChange:  time.Now(),
			},
			"currency_l1": {
				Name:             "Metagraph Currency L1",
				State:            CircuitClosed,
				FailureThreshold: 5,
				SuccessThreshold: 2,
				ResetTimeout:     30 * time.Second,
				LastStateChange:  time.Now(),
			},
			"cardano": {
				Name:             "Cardano",
				State:            CircuitClosed,
				FailureThreshold: 3,
				SuccessThreshold: 2,
				ResetTimeout:     60 * time.Second,
				LastStateChange:  time.Now(),
			},
			"ipfs": {
				Name:             "IPFS/Storj",
				State:            CircuitClosed,
				FailureThreshold: 5,
				SuccessThreshold: 2,
				ResetTimeout:     120 * time.Second,
				LastStateChange:  time.Now(),
			},
			"digital_evidence": {
				Name:             "Digital Evidence API",
				State:            CircuitClosed,
				FailureThreshold: 3,
				SuccessThreshold: 2,
				ResetTimeout:     60 * time.Second,
				LastStateChange:  time.Now(),
			},
		},
	}
}

// AddCircuit registers a new circuit breaker for a downstream dependency.
func (cb *CircuitBreaker) AddCircuit(name, displayName string, failureThreshold, successThreshold int, resetTimeout time.Duration) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.circuits[name] = &Circuit{
		Name:             displayName,
		State:            CircuitClosed,
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		ResetTimeout:     resetTimeout,
		LastStateChange:  time.Now(),
	}
}

// Allow checks if requests should be allowed through for the given circuit.
func (cb *CircuitBreaker) Allow(name string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	c, ok := cb.circuits[name]
	if !ok {
		return true // Unknown circuit — allow by default
	}

	switch c.State {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(c.LastFailure) > c.ResetTimeout {
			c.State = CircuitHalfOpen
			c.SuccessCount = 0
			c.LastStateChange = time.Now()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request for the given circuit.
func (cb *CircuitBreaker) RecordSuccess(name string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	c, ok := cb.circuits[name]
	if !ok {
		return
	}

	switch c.State {
	case CircuitHalfOpen:
		c.SuccessCount++
		if c.SuccessCount >= c.SuccessThreshold {
			c.State = CircuitClosed
			c.FailureCount = 0
			c.SuccessCount = 0
			c.LastStateChange = time.Now()
		}
	case CircuitClosed:
		c.FailureCount = 0 // Reset on success
	}
}

// RecordFailure records a failed request for the given circuit.
func (cb *CircuitBreaker) RecordFailure(name string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	c, ok := cb.circuits[name]
	if !ok {
		return
	}

	c.LastFailure = time.Now()

	switch c.State {
	case CircuitClosed:
		c.FailureCount++
		if c.FailureCount >= c.FailureThreshold {
			c.State = CircuitOpen
			c.LastStateChange = time.Now()
		}
	case CircuitHalfOpen:
		// Any failure in half-open immediately re-opens
		c.State = CircuitOpen
		c.SuccessCount = 0
		c.LastStateChange = time.Now()
	}
}

// GetState returns the current state of a named circuit.
func (cb *CircuitBreaker) GetState(name string) (CircuitState, bool) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	c, ok := cb.circuits[name]
	if !ok {
		return CircuitClosed, false
	}
	return c.State, true
}

// GetAllStates returns a snapshot of all circuit states.
func (cb *CircuitBreaker) GetAllStates() map[string]CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	states := make(map[string]CircuitState, len(cb.circuits))
	for name, c := range cb.circuits {
		states[name] = c.State
	}
	return states
}
