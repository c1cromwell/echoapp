package metagraph

import (
	"testing"
	"time"
)

func TestCircuitBreaker_OpensAfterThresholdAndFastFails(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Hour) // long cooldown so it stays open

	// Below threshold: still closed, calls allowed.
	for i := 0; i < 2; i++ {
		if !cb.Allow() {
			t.Fatalf("call %d should be allowed while closed", i)
		}
		cb.Failure()
	}
	if cb.State() != "closed" {
		t.Fatalf("expected closed after 2 failures, got %s", cb.State())
	}

	// Third failure trips it open.
	if !cb.Allow() {
		t.Fatal("3rd call should still be allowed before the failure trips it")
	}
	cb.Failure()
	if cb.State() != "open" {
		t.Fatalf("expected open after threshold failures, got %s", cb.State())
	}

	// Open → fast-fail without admitting calls.
	if cb.Allow() {
		t.Fatal("breaker open: calls must be rejected")
	}
}

func TestCircuitBreaker_HalfOpenProbeRecovers(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)

	cb.Failure() // trips open (threshold 1)
	if cb.State() != "open" {
		t.Fatalf("expected open, got %s", cb.State())
	}
	if cb.Allow() {
		t.Fatal("should reject during cooldown")
	}

	time.Sleep(15 * time.Millisecond)

	// First Allow after cooldown admits a single probe (half-open)...
	if !cb.Allow() {
		t.Fatal("should admit one probe after cooldown")
	}
	if cb.State() != "half-open" {
		t.Fatalf("expected half-open, got %s", cb.State())
	}
	// ...and concurrent calls are rejected until the probe resolves.
	if cb.Allow() {
		t.Fatal("only one probe allowed in half-open")
	}

	// Probe succeeds → closed again.
	cb.Success()
	if cb.State() != "closed" {
		t.Fatalf("expected closed after successful probe, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("closed breaker should allow calls")
	}
}

func TestCircuitBreaker_HalfOpenProbeFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.Failure()
	time.Sleep(15 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("probe should be admitted")
	}
	cb.Failure() // probe fails → re-open
	if cb.State() != "open" {
		t.Fatalf("expected re-open after failed probe, got %s", cb.State())
	}
}

// guarded should short-circuit with ErrCircuitOpen once the breaker is open, without
// invoking the wrapped function.
func TestGuarded_ShortCircuitsWhenOpen(t *testing.T) {
	c := &MetagraphClient{breaker: NewCircuitBreaker(1, time.Hour)}

	calls := 0
	failing := func() (string, error) { calls++; return "", assertErr }

	if _, err := c.guarded(failing); err != assertErr {
		t.Fatalf("first call should run and return the inner error, got %v", err)
	}
	// Breaker is now open; the next call must not invoke failing.
	_, err := c.guarded(failing)
	if err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("wrapped fn should not run while open; ran %d times", calls)
	}
}

var assertErr = &simpleErr{"boom"}

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }
