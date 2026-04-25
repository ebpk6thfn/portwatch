package portscanner

import (
	"sync"
	"time"
)

// ShadowPolicy controls shadow-mode behaviour.
type ShadowPolicy struct {
	// Enabled puts the pipeline into shadow mode: events are observed but
	// not forwarded to notifiers. Useful for dry-run / canary deployments.
	Enabled bool
	// LogDropped records every suppressed event for later inspection.
	LogDropped bool
	// MaxDropped is the maximum number of dropped events retained in memory.
	// 0 means unlimited.
	MaxDropped int
}

// DefaultShadowPolicy returns a policy with shadow mode disabled.
func DefaultShadowPolicy() ShadowPolicy {
	return ShadowPolicy{
		Enabled:    false,
		LogDropped: true,
		MaxDropped: 512,
	}
}

// DroppedEvent is a ChangeEvent that was suppressed by shadow mode.
type DroppedEvent struct {
	Event     ChangeEvent
	DroppedAt time.Time
}

// Shadow wraps a pipeline stage and, when enabled, swallows events instead
// of forwarding them while optionally retaining them for inspection.
type Shadow struct {
	mu      sync.Mutex
	policy  ShadowPolicy
	dropped []DroppedEvent
}

// NewShadow constructs a Shadow with the given policy.
func NewShadow(p ShadowPolicy) *Shadow {
	return &Shadow{policy: p}
}

// Filter returns nil (suppress) when shadow mode is enabled, otherwise
// returns the event unchanged.
func (s *Shadow) Filter(ev ChangeEvent) *ChangeEvent {
	if !s.policy.Enabled {
		return &ev
	}
	if s.policy.LogDropped {
		s.record(ev)
	}
	return nil
}

func (s *Shadow) record(ev ChangeEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	de := DroppedEvent{Event: ev, DroppedAt: time.Now()}
	if s.policy.MaxDropped > 0 && len(s.dropped) >= s.policy.MaxDropped {
		s.dropped = s.dropped[1:]
	}
	s.dropped = append(s.dropped, de)
}

// Dropped returns a snapshot of all suppressed events.
func (s *Shadow) Dropped() []DroppedEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]DroppedEvent, len(s.dropped))
	copy(out, s.dropped)
	return out
}

// Len returns the number of retained dropped events.
func (s *Shadow) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.dropped)
}

// Clear removes all retained dropped events.
func (s *Shadow) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropped = s.dropped[:0]
}

// IsEnabled reports whether shadow mode is currently active.
func (s *Shadow) IsEnabled() bool { return s.policy.Enabled }
