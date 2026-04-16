package portscanner

import (
	"sync"
	"time"
)

// Debouncer suppresses rapid repeated events for the same key,
// emitting only after a quiet period has elapsed.
type Debouncer struct {
	mu      sync.Mutex
	window  time.Duration
	timers  map[string]time.Time
	now     func() time.Time
}

// NewDebouncer creates a Debouncer with the given quiet window.
func NewDebouncer(window time.Duration) *Debouncer {
	return &Debouncer{
		window: window,
		timers: make(map[string]time.Time),
		now:    time.Now,
	}
}

// Allow returns true if the event for key should be emitted.
// It resets the quiet window on every call; only the final
// event in a burst (after silence) is forwarded.
func (d *Debouncer) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	last, seen := d.timers[key]
	d.timers[key] = now

	if !seen {
		return true
	}
	// Suppress if still within the quiet window.
	if now.Sub(last) < d.window {
		return false
	}
	return true
}

// Flush removes all expired entries older than the window.
func (d *Debouncer) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	for k, t := range d.timers {
		if now.Sub(t) >= d.window {
			delete(d.timers, k)
		}
	}
}

// Len returns the number of tracked keys.
func (d *Debouncer) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.timers)
}
