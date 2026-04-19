package notifier

import (
	"errors"
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("notifier: circuit open, call rejected")

// CircuitNotifier wraps a Notifier with a circuit breaker to stop hammering
// a failing downstream (e.g. webhook endpoint).
type CircuitNotifier struct {
	inner   Notifier
	circuit *portscanner.CircuitBreaker
}

// NewCircuitNotifier wraps inner with a circuit breaker.
// threshold is the number of consecutive failures before opening;
// recoveryWait is how long to wait before probing again.
func NewCircuitNotifier(inner Notifier, threshold int, recoveryWait time.Duration) *CircuitNotifier {
	return &CircuitNotifier{
		inner:   inner,
		circuit: portscanner.NewCircuitBreaker(threshold, recoveryWait),
	}
}

// Notify sends the event through the inner notifier, guarded by the circuit breaker.
func (cn *CircuitNotifier) Notify(event portscanner.ChangeEvent) error {
	if !cn.circuit.Allow() {
		return ErrCircuitOpen
	}
	err := cn.inner.Notify(event)
	if err != nil {
		cn.circuit.RecordFailure()
		return fmt.Errorf("circuit notifier: %w", err)
	}
	cn.circuit.RecordSuccess()
	return nil
}

// State exposes the underlying circuit state for observability.
func (cn *CircuitNotifier) State() portscanner.CircuitState {
	return cn.circuit.State()
}
