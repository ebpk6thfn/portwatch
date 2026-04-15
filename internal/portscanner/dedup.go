package portscanner

import (
	"sync"
	"time"
)

// Deduplicator suppresses duplicate ChangeEvents within a sliding window.
// Two events are considered duplicates if they share the same key, type, and
// occurred within the configured window duration.
type Deduplicator struct {
	mu     sync.Mutex
	seen   map[string]time.Time
	window time.Duration
	now    func() time.Time
}

// NewDeduplicator returns a Deduplicator that suppresses repeated events
// within the given window. Pass a zero duration to disable deduplication.
func NewDeduplicator(window time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:   make(map[string]time.Time),
		window: window,
		now:    time.Now,
	}
}

// IsDuplicate returns true if an equivalent event was already seen within the
// deduplication window. If the event is new (or the window has expired), it
// records the event and returns false.
func (d *Deduplicator) IsDuplicate(event ChangeEvent) bool {
	if d.window <= 0 {
		return false
	}

	key := dedupKey(event)
	now := d.now()

	d.mu.Lock()
	defer d.mu.Unlock()

	if last, ok := d.seen[key]; ok && now.Sub(last) < d.window {
		return true
	}

	d.seen[key] = now
	return false
}

// Filter returns only those events that are not duplicates.
func (d *Deduplicator) Filter(events []ChangeEvent) []ChangeEvent {
	out := make([]ChangeEvent, 0, len(events))
	for _, e := range events {
		if !d.IsDuplicate(e) {
			out = append(out, e)
		}
	}
	return out
}

// Purge removes stale entries older than the window to keep memory bounded.
func (d *Deduplicator) Purge() {
	now := d.now()
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, t := range d.seen {
		if now.Sub(t) >= d.window {
			delete(d.seen, k)
		}
	}
}

func dedupKey(e ChangeEvent) string {
	return e.Entry.Key() + "|" + e.Type
}
