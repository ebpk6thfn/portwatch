package portscanner

import (
	"sync"
	"time"
)

// TagStore persists arbitrary string tags on a per-port-key basis with an
// optional TTL. Tags with a zero expiry never expire.
type TagStore struct {
	mu      sync.RWMutex
	entries map[string]tagEntry
}

type tagEntry struct {
	tags    []string
	expires time.Time // zero means no expiry
}

// NewTagStore returns an empty TagStore.
func NewTagStore() *TagStore {
	return &TagStore{entries: make(map[string]tagEntry)}
}

// Set replaces all tags for key. If ttl is zero the tags never expire.
func (ts *TagStore) Set(key string, tags []string, ttl time.Duration) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	copy := make([]string, len(tags))
	for i, t := range tags {
		copy[i] = t
	}
	ts.entries[key] = tagEntry{tags: copy, expires: exp}
}

// Get returns the tags for key and whether they exist (and have not expired).
func (ts *TagStore) Get(key string) ([]string, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	e, ok := ts.entries[key]
	if !ok {
		return nil, false
	}
	if !e.expires.IsZero() && time.Now().After(e.expires) {
		return nil, false
	}
	out := make([]string, len(e.tags))
	copy(out, e.tags)
	return out, true
}

// Delete removes tags for key.
func (ts *TagStore) Delete(key string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.entries, key)
}

// Flush removes all expired entries.
func (ts *TagStore) Flush() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := time.Now()
	for k, e := range ts.entries {
		if !e.expires.IsZero() && now.After(e.expires) {
			delete(ts.entries, k)
		}
	}
}

// Len returns the number of non-expired entries.
func (ts *TagStore) Len() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	now := time.Now()
	count := 0
	for _, e := range ts.entries {
		if e.expires.IsZero() || !now.After(e.expires) {
			count++
		}
	}
	return count
}
