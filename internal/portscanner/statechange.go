package portscanner

import "time"

// StateChangeTracker records the first and last time each port key was seen
// in a given state (opened/closed), enabling duration-aware alerting.
type StateChangeTracker struct {
	first map[string]time.Time
	last  map[string]time.Time
	now   func() time.Time
}

// NewStateChangeTracker returns a new StateChangeTracker.
func NewStateChangeTracker(now func() time.Time) *StateChangeTracker {
	if now == nil {
		now = time.Now
	}
	return &StateChangeTracker{
		first: make(map[string]time.Time),
		last:  make(map[string]time.Time),
		now:   now,
	}
}

// Record marks the key as active at the current time.
// Returns the first-seen time and how long it has been active.
func (t *StateChangeTracker) Record(key string) (first time.Time, duration time.Duration) {
	now := t.now()
	if _, ok := t.first[key]; !ok {
		t.first[key] = now
	}
	t.last[key] = now
	return t.first[key], now.Sub(t.first[key])
}

// Forget removes the key from tracking.
func (t *StateChangeTracker) Forget(key string) {
	delete(t.first, key)
	delete(t.last, key)
}

// FirstSeen returns the first time the key was recorded, and whether it exists.
func (t *StateChangeTracker) FirstSeen(key string) (time.Time, bool) {
	v, ok := t.first[key]
	return v, ok
}

// LastSeen returns the last time the key was recorded, and whether it exists.
func (t *StateChangeTracker) LastSeen(key string) (time.Time, bool) {
	v, ok := t.last[key]
	return v, ok
}

// Len returns the number of tracked keys.
func (t *StateChangeTracker) Len() int {
	return len(t.first)
}
