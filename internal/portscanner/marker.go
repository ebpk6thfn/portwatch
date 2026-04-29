package portscanner

import (
	"sync"
	"time"
)

// MarkerPolicy controls how long a mark persists.
type MarkerPolicy struct {
	TTL time.Duration
}

// DefaultMarkerPolicy returns a sensible default.
func DefaultMarkerPolicy() MarkerPolicy {
	return MarkerPolicy{
		TTL: 10 * time.Minute,
	}
}

type markerEntry struct {
	label    string
	markedAt time.Time
}

// Marker allows tagging port-event keys with an arbitrary label for a
// configurable duration. Useful for annotating events that need manual
// follow-up or suppression by an operator.
type Marker struct {
	mu     sync.Mutex
	policy MarkerPolicy
	now    func() time.Time
	marks  map[string]markerEntry
}

// NewMarker creates a Marker with the given policy.
func NewMarker(policy MarkerPolicy) *Marker {
	return &Marker{
		policy: policy,
		now:    time.Now,
		marks:  make(map[string]markerEntry),
	}
}

// Mark associates a label with key for the policy TTL.
func (m *Marker) Mark(key, label string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.marks[key] = markerEntry{label: label, markedAt: m.now()}
}

// Unmark removes any mark for key.
func (m *Marker) Unmark(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.marks, key)
}

// Get returns the label for key and whether it is still active.
func (m *Marker) Get(key string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.marks[key]
	if !ok {
		return "", false
	}
	if m.policy.TTL > 0 && m.now().Sub(e.markedAt) > m.policy.TTL {
		delete(m.marks, key)
		return "", false
	}
	return e.label, true
}

// Flush removes all expired marks.
func (m *Marker) Flush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	for k, e := range m.marks {
		if m.policy.TTL > 0 && now.Sub(e.markedAt) > m.policy.TTL {
			delete(m.marks, k)
		}
	}
}

// Len returns the number of currently stored (possibly expired) marks.
func (m *Marker) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.marks)
}
