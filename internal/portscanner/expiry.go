package portscanner

import (
	"sync"
	"time"
)

// ExpiryPolicy controls how the ExpiryTracker behaves.
type ExpiryPolicy struct {
	// TTL is how long a key is considered alive after its last touch.
	TTL time.Duration
}

// DefaultExpiryPolicy returns a sensible default policy.
func DefaultExpiryPolicy() ExpiryPolicy {
	return ExpiryPolicy{
		TTL: 5 * time.Minute,
	}
}

// ExpiryEntry holds metadata for a tracked key.
type ExpiryEntry struct {
	Key       string
	FirstSeen time.Time
	LastSeen  time.Time
	ExpiredAt time.Time
}

// IsExpired reports whether the entry has passed its TTL relative to now.
func (e ExpiryEntry) IsExpired(now time.Time, ttl time.Duration) bool {
	return now.After(e.LastSeen.Add(ttl))
}

// ExpiryTracker records when keys were first and last seen, and detects
// when they have not been observed within the configured TTL.
type ExpiryTracker struct {
	mu     sync.Mutex
	policy ExpiryPolicy
	now    func() time.Time
	entries map[string]*ExpiryEntry
}

// NewExpiryTracker creates an ExpiryTracker with the given policy.
// If nowFn is nil, time.Now is used.
func NewExpiryTracker(policy ExpiryPolicy, nowFn func() time.Time) *ExpiryTracker {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &ExpiryTracker{
		policy:  policy,
		now:     nowFn,
		entries: make(map[string]*ExpiryEntry),
	}
}

// Touch records an observation of key at the current time.
// Returns the entry after updating it.
func (t *ExpiryTracker) Touch(key string) ExpiryEntry {
	now := t.now()
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	if !ok {
		e = &ExpiryEntry{Key: key, FirstSeen: now}
		t.entries[key] = e
	}
	e.LastSeen = now
	return *e
}

// Expired returns all entries whose LastSeen is older than TTL and removes
// them from the tracker, stamping ExpiredAt with the current time.
func (t *ExpiryTracker) Expired() []ExpiryEntry {
	now := t.now()
	t.mu.Lock()
	defer t.mu.Unlock()
	var out []ExpiryEntry
	for k, e := range t.entries {
		if e.IsExpired(now, t.policy.TTL) {
			e.ExpiredAt = now
			out = append(out, *e)
			delete(t.entries, k)
		}
	}
	return out
}

// Len returns the number of currently tracked (non-expired) keys.
func (t *ExpiryTracker) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}
