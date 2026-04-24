package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// EscalationPolicy defines thresholds for escalating event severity.
type EscalationPolicy struct {
	// If the same port fires more than CountThreshold events within Window,
	// its severity is escalated to High.
	CountThreshold int
	Window         time.Duration
}

// DefaultEscalationPolicy returns a sensible default escalation policy.
func DefaultEscalationPolicy() EscalationPolicy {
	return EscalationPolicy{
		CountThreshold: 3,
		Window:         2 * time.Minute,
	}
}

// escalationEntry tracks recent event timestamps for a key.
type escalationEntry struct {
	times []time.Time
}

// Escalator promotes event severity when a port exceeds a burst threshold.
type Escalator struct {
	mu     sync.Mutex
	policy EscalationPolicy
	record map[string]*escalationEntry
	now    func() time.Time
}

// NewEscalator creates an Escalator with the given policy.
func NewEscalator(policy EscalationPolicy) *Escalator {
	return &Escalator{
		policy: policy,
		record: make(map[string]*escalationEntry),
		now:    time.Now,
	}
}

// Process checks whether the event should be escalated and returns a
// (possibly modified) copy of the event.
func (e *Escalator) Process(ev ChangeEvent) ChangeEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%d", ev.Entry.Protocol, ev.Entry.Port)
	now := e.now()
	cutoff := now.Add(-e.policy.Window)

	ent, ok := e.record[key]
	if !ok {
		ent = &escalationEntry{}
		e.record[key] = ent
	}

	// Evict old timestamps.
	filtered := ent.times[:0]
	for _, t := range ent.times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	ent.times = append(filtered, now)

	if len(ent.times) >= e.policy.CountThreshold {
		ev.Severity = SeverityHigh
	}
	return ev
}

// Flush removes all tracking state older than the policy window.
func (e *Escalator) Flush() {
	e.mu.Lock()
	defer e.mu.Unlock()

	cutoff := e.now().Add(-e.policy.Window)
	for key, ent := range e.record {
		filtered := ent.times[:0]
		for _, t := range ent.times {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 0 {
			delete(e.record, key)
		} else {
			ent.times = filtered
		}
	}
}
