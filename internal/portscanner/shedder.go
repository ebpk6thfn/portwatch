package portscanner

import (
	"sync"
	"time"
)

// ShedderPolicy controls load-shedding behaviour.
type ShedderPolicy struct {
	// MaxQueueDepth is the maximum number of pending events before shedding begins.
	MaxQueueDepth int
	// ShedPercent is the fraction of events to drop (0.0–1.0) when overloaded.
	ShedPercent float64
	// CooldownPeriod is how long to remain in shedding mode after the queue drains.
	CooldownPeriod time.Duration
}

// DefaultShedderPolicy returns a sensible default load-shedding policy.
func DefaultShedderPolicy() ShedderPolicy {
	return ShedderPolicy{
		MaxQueueDepth:  200,
		ShedPercent:    0.5,
		CooldownPeriod: 10 * time.Second,
	}
}

// Shedder drops a configurable fraction of events when the queue depth
// exceeds a high-watermark, protecting downstream consumers from overload.
type Shedder struct {
	mu       sync.Mutex
	policy   ShedderPolicy
	depth    int
	shedding bool
	coolUntil time.Time
	counter  int
	now      func() time.Time
}

// NewShedder creates a Shedder with the given policy.
func NewShedder(p ShedderPolicy) *Shedder {
	return &Shedder{policy: p, now: time.Now}
}

// SetDepth updates the observed queue depth. Call this before Allow.
func (s *Shedder) SetDepth(d int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.depth = d
}

// Allow returns true if the event should be forwarded.
// When depth exceeds MaxQueueDepth the shedder enters shedding mode and
// drops ShedPercent of events in round-robin fashion.
func (s *Shedder) Allow(e ChangeEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()

	// Activate shedding when depth is too high.
	if s.depth >= s.policy.MaxQueueDepth {
		s.shedding = true
		s.coolUntil = now.Add(s.policy.CooldownPeriod)
	}

	// Deactivate after cooldown elapses.
	if s.shedding && now.After(s.coolUntil) {
		s.shedding = false
		s.counter = 0
	}

	if !s.shedding {
		return true
	}

	// Round-robin drop: keep 1 every (1/keepRatio) events.
	s.counter++
	keepEvery := 1
	if s.policy.ShedPercent > 0 && s.policy.ShedPercent < 1.0 {
		keepEvery = int(1.0 / (1.0 - s.policy.ShedPercent))
		if keepEvery < 1 {
			keepEvery = 1
		}
	} else if s.policy.ShedPercent >= 1.0 {
		return false
	}
	return s.counter%keepEvery == 0
}

// IsShedding reports whether the shedder is currently active.
func (s *Shedder) IsShedding() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shedding
}

// Filter returns only the events that pass the shedder.
func (s *Shedder) Filter(events []ChangeEvent) []ChangeEvent {
	out := events[:0:0]
	for _, e := range events {
		if s.Allow(e) {
			out = append(out, e)
		}
	}
	return out
}
