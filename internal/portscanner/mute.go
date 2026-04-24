package portscanner

import (
	"sync"
	"time"
)

// MutePolicy controls how long a mute window lasts.
type MutePolicy struct {
	Duration time.Duration
}

// DefaultMutePolicy returns a sensible default mute policy.
func DefaultMutePolicy() MutePolicy {
	return MutePolicy{
		Duration: 30 * time.Minute,
	}
}

// muteEntry tracks when a key was muted and when it expires.
type muteEntry struct {
	mutedAt time.Time
	expiresAt time.Time
}

// Muter suppresses events for a key until the mute window expires.
type Muter struct {
	mu      sync.Mutex
	policy  MutePolicy
	entries map[string]muteEntry
	now     func() time.Time
}

// NewMuter creates a Muter with the given policy.
func NewMuter(policy MutePolicy) *Muter {
	return &Muter{
		policy:  policy,
		entries: make(map[string]muteEntry),
		now:     time.Now,
	}
}

// Mute silences events for the given key for the policy duration.
func (m *Muter) Mute(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	m.entries[key] = muteEntry{
		mutedAt:   now,
		expiresAt: now.Add(m.policy.Duration),
	}
}

// Unmute removes the mute for the given key immediately.
func (m *Muter) Unmute(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, key)
}

// IsMuted returns true if the key is currently muted.
func (m *Muter) IsMuted(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[key]
	if !ok {
		return false
	}
	if m.now().After(e.expiresAt) {
		delete(m.entries, key)
		return false
	}
	return true
}

// Filter returns only the events whose key is not currently muted.
// The key is derived from event.Entry.Key().
func (m *Muter) Filter(events []ChangeEvent) []ChangeEvent {
	out := events[:0:0]
	for _, ev := range events {
		if !m.IsMuted(ev.Entry.Key()) {
			out = append(out, ev)
		}
	}
	return out
}
