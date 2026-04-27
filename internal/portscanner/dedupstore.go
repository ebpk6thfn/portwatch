package portscanner

import (
	"sync"
	"time"
)

// DedupStore tracks recently seen event keys with expiry to prevent
// duplicate notifications across daemon restarts via persistent-style memory.
type DedupStore struct {
	mu      sync.Mutex
	window  time.Duration
	entries map[string]time.Time
	now     func() time.Time
}

// NewDedupStore creates a DedupStore with the given deduplication window.
func NewDedupStore(window time.Duration) *DedupStore {
	return &DedupStore{
		window:  window,
		entries: make(map[string]time.Time),
		now:     time.Now,
	}
}

// Seen returns true if the key was recorded within the dedup window.
// If not seen (or expired), it records the key and returns false.
func (d *DedupStore) Seen(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	if t, ok := d.entries[key]; ok && now.Sub(t) < d.window {
		return true
	}
	d.entries[key] = now
	return false
}

// Flush removes all entries that have exceeded the dedup window.
func (d *DedupStore) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	for k, t := range d.entries {
		if now.Sub(t) >= d.window {
			delete(d.entries, k)
		}
	}
}

// Len returns the number of active (non-expired) entries.
func (d *DedupStore) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	count := 0
	for _, t := range d.entries {
		if now.Sub(t) < d.window {
			count++
		}
	}
	return count
}

// Delete removes a key from the store immediately, regardless of its expiry.
// This is useful when an event is explicitly invalidated (e.g. a port closes).
func (d *DedupStore) Delete(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, key)
}
