package portscanner

import (
	"testing"
	"time"
)

func fixedCircuitNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestCircuit_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Second)
	if cb.State() != CircuitClosed {
		t.Fatal("expected closed")
	}
	if !cb.Allow() {
		t.Fatal("expected allow")
	}
}

func TestCircuit_TripsAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitClosed {
		t.Fatal("should still be closed")
	}
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("expected open after threshold")
	}
	if cb.Allow() {
		t.Fatal("open circuit should reject")
	}
}

func TestCircuit_HalfOpenAfterRecovery(t *testing.T) {
	base := time.Now()
	cb := NewCircuitBreaker(1, 5*time.Second)
	cb.now = fixedCircuitNow(base)
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("expected open")
	}
	cb.now = fixedCircuitNow(base.Add(6 * time.Second))
	if !cb.Allow() {
		t.Fatal("should allow in half-open")
	}
	if cb.State() != CircuitHalfOpen {
		t.Fatal("expected half-open")
	}
}

func TestCircuit_RecoveryClosesBreaker(t *testing.T) {
	cb := NewCircuitBreaker(1, time.Second)
	cb.RecordFailure()
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatal("expected closed after success")
	}
	if !cb.Allow() {
		t.Fatal("closed should allow")
	}
}

func TestCircuit_SuccessResetsFailureCount(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()
	cb.RecordFailure()
	if cb.State() != CircuitClosed {
		t.Fatal("failure count should have reset")
	}
}
