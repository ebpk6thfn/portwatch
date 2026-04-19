package portscanner

import (
	"math"
	"sync"
	"time"
)

// BackoffStrategy defines how retry delays are calculated.
type BackoffStrategy int

const (
	BackoffLinear BackoffStrategy = iota
	BackoffExponential
)

// BackoffPolicy holds configuration for a backoff tracker.
type BackoffPolicy struct {
	Strategy  BackoffStrategy
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Multipler float64
}

// DefaultBackoffPolicy returns a sensible exponential backoff policy.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		Strategy:  BackoffExponential,
		BaseDelay: 1 * time.Second,
		MaxDelay:  60 * time.Second,
		Multipler: 2.0,
	}
}

// Backoff tracks per-key retry attempt counts and computes delays.
type Backoff struct {
	mu       sync.Mutex
	policy   BackoffPolicy
	attempts map[string]int
}

// NewBackoff creates a new Backoff tracker with the given policy.
func NewBackoff(policy BackoffPolicy) *Backoff {
	return &Backoff{
		policy:   policy,
		attempts: make(map[string]int),
	}
}

// Record increments the attempt count for the given key and returns the delay
// that should be waited before the next retry.
func (b *Backoff) Record(key string) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.attempts[key]++
	attempt := b.attempts[key]

	var delay time.Duration
	switch b.policy.Strategy {
	case BackoffExponential:
		factor := math.Pow(b.policy.Multipler, float64(attempt-1))
		delay = time.Duration(float64(b.policy.BaseDelay) * factor)
	default:
		delay = b.policy.BaseDelay * time.Duration(attempt)
	}

	if delay > b.policy.MaxDelay {
		delay = b.policy.MaxDelay
	}
	return delay
}

// Reset clears the attempt count for the given key.
func (b *Backoff) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.attempts, key)
}

// Attempts returns the current attempt count for a key.
func (b *Backoff) Attempts(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.attempts[key]
}
