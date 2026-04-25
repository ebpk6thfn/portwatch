package portscanner

import (
	"sync"
	"time"
)

// FencePolicy defines thresholds for the rate-limit fence.
type FencePolicy struct {
	// MaxEvents is the maximum number of events allowed within Window before
	// the fence activates.
	MaxEvents int
	// Window is the rolling time window used for counting.
	Window time.Duration
	// CooldownAfterFence is how long the fence stays active once tripped.
	CooldownAfterFence time.Duration
}

// DefaultFencePolicy returns a sensible default FencePolicy.
func DefaultFencePolicy() FencePolicy {
	return FencePolicy{
		MaxEvents:          50,
		Window:             time.Minute,
		CooldownAfterFence: 2 * time.Minute,
	}
}

// Fence is a global rate-limit gate that blocks all events once the total
// event rate across all keys exceeds a configured threshold. It resets
// automatically after a cooldown period.
type Fence struct {
	mu         sync.Mutex
	policy     FencePolicy
	events     []time.Time
	fencedAt   time.Time
	isFenced   bool
	now        func() time.Time
}

// NewFence constructs a Fence with the given policy.
func NewFence(policy FencePolicy) *Fence {
	return &Fence{
		policy: policy,
		now:    time.Now,
	}
}

// Allow returns true if the event should be allowed through, false if the
// fence is active. Each call to Allow records an event timestamp.
func (f *Fence) Allow() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := f.now()

	// If currently fenced, check whether the cooldown has elapsed.
	if f.isFenced {
		if now.Before(f.fencedAt.Add(f.policy.CooldownAfterFence)) {
			return false
		}
		// Cooldown elapsed — lift the fence.
		f.isFenced = false
		f.events = f.events[:0]
	}

	// Evict events outside the rolling window.
	cutoff := now.Add(-f.policy.Window)
	valid := f.events[:0]
	for _, t := range f.events {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	f.events = valid

	// Record this event.
	f.events = append(f.events, now)

	// Trip the fence if threshold exceeded.
	if len(f.events) > f.policy.MaxEvents {
		f.isFenced = true
		f.fencedAt = now
		return false
	}

	return true
}

// IsFenced reports whether the fence is currently active.
func (f *Fence) IsFenced() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.isFenced
}

// Reset clears all state, lifting any active fence immediately.
func (f *Fence) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = f.events[:0]
	f.isFenced = false
	f.fencedAt = time.Time{}
}
