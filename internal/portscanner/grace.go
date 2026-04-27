package portscanner

import (
	"sync"
	"time"
)

// GracePolicy defines timing parameters for the grace period filter.
type GracePolicy struct {
	// Window is how long after startup events are suppressed.
	Window time.Duration
}

// DefaultGracePolicy returns a sensible default: suppress events for 5 seconds
// after startup to avoid alert storms on daemon restart.
func DefaultGracePolicy() GracePolicy {
	return GracePolicy{
		Window: 5 * time.Second,
	}
}

// Grace suppresses all ChangeEvents that occur within a startup window.
// This prevents alert storms caused by the initial port scan on daemon start.
type Grace struct {
	policy    GracePolicy
	startedAt time.Time
	now       func() time.Time
	mu        sync.Mutex
}

// NewGrace creates a Grace filter that begins its window at construction time.
func NewGrace(policy GracePolicy) *Grace {
	return newGraceWithClock(policy, time.Now)
}

func newGraceWithClock(policy GracePolicy, now func() time.Time) *Grace {
	return &Grace{
		policy:    policy,
		startedAt: now(),
		now:       now,
	}
}

// Allow returns true if the event should be forwarded (grace window has elapsed).
func (g *Grace) Allow(_ ChangeEvent) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.now().Sub(g.startedAt) >= g.policy.Window
}

// Filter returns only the events that fall outside the grace window.
func (g *Grace) Filter(events []ChangeEvent) []ChangeEvent {
	out := events[:0]
	for _, e := range events {
		if g.Allow(e) {
			out = append(out, e)
		}
	}
	return out
}

// Elapsed returns the time elapsed since startup.
func (g *Grace) Elapsed() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.now().Sub(g.startedAt)
}

// InWindow reports whether the grace window is still active.
func (g *Grace) InWindow() bool {
	return !g.Allow(ChangeEvent{})
}
