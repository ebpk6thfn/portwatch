package portscanner

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal operation
	CircuitOpen                        // failing, rejecting calls
	CircuitHalfOpen                    // probing for recovery
)

// CircuitBreaker trips open after consecutive failures and recovers after a timeout.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	recoveryWait time.Duration
	openedAt     time.Time
	now          func() time.Time
}

// NewCircuitBreaker creates a CircuitBreaker with the given failure threshold and recovery wait.
func NewCircuitBreaker(threshold int, recoveryWait time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		recoveryWait: recoveryWait,
		now:          time.Now,
	}
}

// Allow returns true if the call should proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if cb.now().Sub(cb.openedAt) >= cb.recoveryWait {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess resets the breaker to closed.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure increments the failure count and may trip the breaker open.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
		cb.openedAt = cb.now()
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
