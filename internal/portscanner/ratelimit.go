package portscanner

import (
	"sync"
	"time"
)

// RateLimiter suppresses repeated change events for the same port key
// within a configurable cooldown window. This prevents alert storms when
// a port flaps open and closed in rapid succession.
type RateLimiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	seen     map[string]time.Time
	now      func() time.Time
}

// NewRateLimiter creates a RateLimiter with the given cooldown duration.
// Events for the same key within the cooldown window are suppressed.
func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		cooldown: cooldown,
		seen:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if the event for the given key should be forwarded,
// or false if it falls within the cooldown window of a prior event.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	if last, ok := r.seen[key]; ok {
		if now.Sub(last) < r.cooldown {
			return false
		}
	}
	r.seen[key] = now
	return true
}

// Filter returns only the ChangeEvents that pass the rate limit.
func (r *RateLimiter) Filter(events []ChangeEvent) []ChangeEvent {
	filtered := make([]ChangeEvent, 0, len(events))
	for _, e := range events {
		if r.Allow(e.Entry.Key()) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Purge removes stale entries older than the cooldown to prevent unbounded growth.
func (r *RateLimiter) Purge() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	for k, t := range r.seen {
		if now.Sub(t) >= r.cooldown {
			delete(r.seen, k)
		}
	}
}
