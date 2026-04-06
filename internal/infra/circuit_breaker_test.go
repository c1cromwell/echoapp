package infra

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker_DefaultCircuits(t *testing.T) {
	cb := NewCircuitBreaker()
	expected := []string{"data_l1", "currency_l1", "cardano", "ipfs", "digital_evidence"}
	for _, name := range expected {
		state, ok := cb.GetState(name)
		if !ok {
			t.Errorf("expected circuit %s to exist", name)
			continue
		}
		if state != CircuitClosed {
			t.Errorf("circuit %s should start closed, got %s", name, state)
		}
	}
}

func TestCircuitBreaker_AllowWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker()
	if !cb.Allow("data_l1") {
		t.Error("closed circuit should allow requests")
	}
}

func TestCircuitBreaker_UnknownCircuitAllows(t *testing.T) {
	cb := NewCircuitBreaker()
	if !cb.Allow("nonexistent") {
		t.Error("unknown circuit should allow by default")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker()

	// data_l1 has threshold of 5
	for i := 0; i < 5; i++ {
		cb.RecordFailure("data_l1")
	}

	state, _ := cb.GetState("data_l1")
	if state != CircuitOpen {
		t.Errorf("circuit should be open after %d failures, got %s", 5, state)
	}

	if cb.Allow("data_l1") {
		t.Error("open circuit should not allow requests")
	}
}

func TestCircuitBreaker_CardanoLowerThreshold(t *testing.T) {
	cb := NewCircuitBreaker()

	// Cardano has threshold of 3
	for i := 0; i < 3; i++ {
		cb.RecordFailure("cardano")
	}

	state, _ := cb.GetState("cardano")
	if state != CircuitOpen {
		t.Errorf("cardano circuit should open after 3 failures, got %s", state)
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker()

	// Use a custom circuit with short timeout for testing
	cb.AddCircuit("test_svc", "Test Service", 2, 1, 10*time.Millisecond)

	cb.RecordFailure("test_svc")
	cb.RecordFailure("test_svc")

	state, _ := cb.GetState("test_svc")
	if state != CircuitOpen {
		t.Fatalf("expected open, got %s", state)
	}

	// Wait for reset timeout
	time.Sleep(15 * time.Millisecond)

	// Should transition to half-open on Allow check
	if !cb.Allow("test_svc") {
		t.Error("should allow after reset timeout (half-open)")
	}

	state, _ = cb.GetState("test_svc")
	if state != CircuitHalfOpen {
		t.Errorf("expected half_open, got %s", state)
	}
}

func TestCircuitBreaker_ClosesAfterSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker()
	cb.AddCircuit("test_svc", "Test", 1, 2, 10*time.Millisecond)

	// Open the circuit
	cb.RecordFailure("test_svc")
	time.Sleep(15 * time.Millisecond)
	cb.Allow("test_svc") // transitions to half-open

	// Successes in half-open close the circuit
	cb.RecordSuccess("test_svc")
	cb.RecordSuccess("test_svc")

	state, _ := cb.GetState("test_svc")
	if state != CircuitClosed {
		t.Errorf("expected closed after successes in half-open, got %s", state)
	}
}

func TestCircuitBreaker_FailureInHalfOpenReopens(t *testing.T) {
	cb := NewCircuitBreaker()
	cb.AddCircuit("test_svc", "Test", 1, 2, 10*time.Millisecond)

	cb.RecordFailure("test_svc")
	time.Sleep(15 * time.Millisecond)
	cb.Allow("test_svc") // half-open

	// Failure in half-open reopens
	cb.RecordFailure("test_svc")

	state, _ := cb.GetState("test_svc")
	if state != CircuitOpen {
		t.Errorf("expected open after failure in half-open, got %s", state)
	}
}

func TestCircuitBreaker_SuccessResetsClosed(t *testing.T) {
	cb := NewCircuitBreaker()

	// Record 4 failures (below threshold of 5)
	for i := 0; i < 4; i++ {
		cb.RecordFailure("data_l1")
	}

	// Success resets the count
	cb.RecordSuccess("data_l1")

	// Now 4 more failures should not open (count was reset)
	for i := 0; i < 4; i++ {
		cb.RecordFailure("data_l1")
	}

	state, _ := cb.GetState("data_l1")
	if state != CircuitClosed {
		t.Errorf("expected closed (count reset by success), got %s", state)
	}
}

func TestCircuitBreaker_GetAllStates(t *testing.T) {
	cb := NewCircuitBreaker()
	states := cb.GetAllStates()

	if len(states) != 5 {
		t.Errorf("expected 5 circuits, got %d", len(states))
	}

	for name, state := range states {
		if state != CircuitClosed {
			t.Errorf("circuit %s should start closed", name)
		}
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := map[CircuitState]string{
		CircuitClosed:   "closed",
		CircuitOpen:     "open",
		CircuitHalfOpen: "half_open",
		CircuitState(99): "unknown",
	}
	for state, expected := range tests {
		if state.String() != expected {
			t.Errorf("state %d: expected %s, got %s", state, expected, state.String())
		}
	}
}
