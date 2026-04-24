package portscanner

import (
	"sync"
	"time"
)

// BackpressurePolicy controls how the backpressure limiter behaves.
type BackpressurePolicy struct {
	// HighWatermark is the queue depth at which backpressure is applied.
	HighWatermark int
	// LowWatermark is the queue depth at which backpressure is released.
	LowWatermark int
	// CooldownPeriod is the minimum time to stay in backpressure state.
	CooldownPeriod time.Duration
}

// DefaultBackpressurePolicy returns a sensible default policy.
func DefaultBackpressurePolicy() BackpressurePolicy {
	return BackpressurePolicy{
		HighWatermark:  100,
		LowWatermark:   25,
		CooldownPeriod: 5 * time.Second,
	}
}

// Backpressure tracks queue depth and signals when the pipeline should slow down.
type Backpressure struct {
	mu       sync.Mutex
	policy   BackpressurePolicy
	depth    int
	active   bool
	enteredAt time.Time
	now      func() time.Time
}

// NewBackpressure creates a new Backpressure limiter with the given policy.
func NewBackpressure(policy BackpressurePolicy, now func() time.Time) *Backpressure {
	if now == nil {
		now = time.Now
	}
	return &Backpressure{policy: policy, now: now}
}

// Push records that one item has been added to the queue.
// Returns true if backpressure is currently active.
func (b *Backpressure) Push() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.depth++
	if !b.active && b.depth >= b.policy.HighWatermark {
		b.active = true
		b.enteredAt = b.now()
	}
	return b.active
}

// Pop records that one item has been removed from the queue.
// Returns true if backpressure is currently active.
func (b *Backpressure) Pop() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.depth > 0 {
		b.depth--
	}
	if b.active {
		cooldownElapsed := b.now().Sub(b.enteredAt) >= b.policy.CooldownPeriod
		if b.depth <= b.policy.LowWatermark && cooldownElapsed {
			b.active = false
		}
	}
	return b.active
}

// IsActive returns whether backpressure is currently engaged.
func (b *Backpressure) IsActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// Depth returns the current queue depth.
func (b *Backpressure) Depth() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.depth
}
