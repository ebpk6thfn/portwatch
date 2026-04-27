package portscanner

import (
	"sync"
	"time"
)

// InhibitPolicy defines configuration for the Inhibitor.
type InhibitPolicy struct {
	// Duration is how long an inhibit rule remains active.
	Duration time.Duration
}

// DefaultInhibitPolicy returns a sensible default inhibit policy.
func DefaultInhibitPolicy() InhibitPolicy {
	return InhibitPolicy{
		Duration: 5 * time.Minute,
	}
}

// inhibitEntry tracks when an inhibit was set and how long it lasts.
type inhibitEntry struct {
	setAt    time.Time
	duration time.Duration
}

func (e inhibitEntry) active(now time.Time) bool {
	return now.Before(e.setAt.Add(e.duration))
}

// Inhibitor temporarily suppresses events for a specific key.
// Unlike Muter (which is externally driven), Inhibitor is triggered
// automatically when a matching condition is met during processing.
type Inhibitor struct {
	mu     sync.Mutex
	rules  map[string]inhibitEntry
	policy InhibitPolicy
	clock  func() time.Time
}

// NewInhibitor creates a new Inhibitor with the given policy.
func NewInhibitor(policy InhibitPolicy) *Inhibitor {
	return &Inhibitor{
		rules:  make(map[string]inhibitEntry),
		policy: policy,
		clock:  time.Now,
	}
}

// Inhibit marks the given key as inhibited for the policy duration.
func (i *Inhibitor) Inhibit(key string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.rules[key] = inhibitEntry{
		setAt:    i.clock(),
		duration: i.policy.Duration,
	}
}

// IsInhibited returns true if the key is currently inhibited.
func (i *Inhibitor) IsInhibited(key string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	entry, ok := i.rules[key]
	if !ok {
		return false
	}
	if !entry.active(i.clock()) {
		delete(i.rules, key)
		return false
	}
	return true
}

// Release removes an inhibit rule for the given key immediately.
func (i *Inhibitor) Release(key string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.rules, key)
}

// Filter returns only events whose keys are not currently inhibited.
func (i *Inhibitor) Filter(events []ChangeEvent) []ChangeEvent {
	out := events[:0:0]
	for _, ev := range events {
		if !i.IsInhibited(ev.Entry.Key()) {
			out = append(out, ev)
		}
	}
	return out
}

// Flush removes all expired inhibit rules.
func (i *Inhibitor) Flush() {
	i.mu.Lock()
	defer i.mu.Unlock()
	now := i.clock()
	for k, e := range i.rules {
		if !e.active(now) {
			delete(i.rules, k)
		}
	}
}
