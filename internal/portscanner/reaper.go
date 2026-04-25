package portscanner

import (
	"sync"
	"time"
)

// ReaperPolicy controls how stale entries are evicted.
type ReaperPolicy struct {
	MaxAge   time.Duration
	Interval time.Duration
}

// DefaultReaperPolicy returns sensible defaults.
func DefaultReaperPolicy() ReaperPolicy {
	return ReaperPolicy{
		MaxAge:   10 * time.Minute,
		Interval: 2 * time.Minute,
	}
}

// ReaperEntry holds a value with a timestamp.
type ReaperEntry struct {
	Key       string
	LastSeen  time.Time
	EventCount int
}

// Reaper periodically evicts entries that have not been seen within MaxAge.
type Reaper struct {
	mu      sync.Mutex
	policy  ReaperPolicy
	entries map[string]*ReaperEntry
	now     func() time.Time
}

// NewReaper constructs a Reaper with the given policy.
func NewReaper(policy ReaperPolicy) *Reaper {
	return &Reaper{
		policy:  policy,
		entries: make(map[string]*ReaperEntry),
		now:     time.Now,
	}
}

// Touch records or refreshes an entry by key.
func (r *Reaper) Touch(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.entries[key]; ok {
		e.LastSeen = r.now()
		e.EventCount++
		return
	}
	r.entries[key] = &ReaperEntry{Key: key, LastSeen: r.now(), EventCount: 1}
}

// Reap removes entries older than MaxAge and returns the reaped keys.
func (r *Reaper) Reap() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := r.now().Add(-r.policy.MaxAge)
	var reaped []string
	for k, e := range r.entries {
		if e.LastSeen.Before(cutoff) {
			reaped = append(reaped, k)
			delete(r.entries, k)
		}
	}
	return reaped
}

// Len returns the current number of tracked entries.
func (r *Reaper) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}

// Get returns the entry for a key, if present.
func (r *Reaper) Get(key string) (*ReaperEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[key]
	if !ok {
		return nil, false
	}
	copy := *e
	return &copy, true
}
