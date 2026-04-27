package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// FlapPolicy controls flap detection behaviour.
type FlapPolicy struct {
	// Window is the duration over which state changes are counted.
	Window time.Duration
	// Threshold is the number of open/close cycles within Window to consider flapping.
	Threshold int
	// Cooldown is how long to suppress alerts after flap is detected.
	Cooldown time.Duration
}

// DefaultFlapPolicy returns sensible defaults.
func DefaultFlapPolicy() FlapPolicy {
	return FlapPolicy{
		Window:    2 * time.Minute,
		Threshold: 3,
		Cooldown:  5 * time.Minute,
	}
}

type flapEntry struct {
	transitions []time.Time
	suppressedUntil time.Time
}

// FlapDetector detects ports that rapidly open and close (flapping).
type FlapDetector struct {
	mu     sync.Mutex
	policy FlapPolicy
	state  map[string]*flapEntry
	now    func() time.Time
}

// NewFlapDetector creates a FlapDetector with the given policy.
func NewFlapDetector(policy FlapPolicy) *FlapDetector {
	return &FlapDetector{
		policy: policy,
		state:  make(map[string]*flapEntry),
		now:    time.Now,
	}
}

// Record records a state change event and returns true if the port is flapping.
func (f *FlapDetector) Record(e ChangeEvent) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := f.now()
	key := e.Entry.Key()

	ent, ok := f.state[key]
	if !ok {
		ent = &flapEntry{}
		f.state[key] = ent
	}

	// If still in cooldown, suppress without recording.
	if now.Before(ent.suppressedUntil) {
		return true
	}

	// Evict old transitions outside the window.
	cutoff := now.Add(-f.policy.Window)
	filtered := ent.transitions[:0]
	for _, t := range ent.transitions {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	ent.transitions = append(filtered, now)

	if len(ent.transitions) >= f.policy.Threshold {
		ent.suppressedUntil = now.Add(f.policy.Cooldown)
		ent.transitions = nil
		return true
	}
	return false
}

// Count returns the current transition count for a key within the window.
func (f *FlapDetector) Count(key string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	ent, ok := f.state[key]
	if !ok {
		return 0
	}
	return len(ent.transitions)
}

// String returns a human-readable description of the detector state.
func (f *FlapDetector) String() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return fmt.Sprintf("FlapDetector{tracked=%d, threshold=%d, window=%s}",
		len(f.state), f.policy.Threshold, f.policy.Window)
}
