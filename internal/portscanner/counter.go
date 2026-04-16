package portscanner

import (
	"sync"
	"time"
)

// Counter tracks how many times each port key has fired an event
// within a rolling time window.
type Counter struct {
	mu      sync.Mutex
	window  time.Duration
	entries map[string][]time.Time
	now     func() time.Time
}

// NewCounter creates a Counter with the given rolling window.
func NewCounter(window time.Duration) *Counter {
	return &Counter{
		window:  window,
		entries: make(map[string][]time.Time),
		now:     time.Now,
	}
}

// Add records an occurrence for key and returns the updated count
// within the current window.
func (c *Counter) Add(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	c.entries[key] = append(c.entries[key], now)
	c.evict(key, now)
	return len(c.entries[key])
}

// Count returns the number of occurrences for key within the window
// without recording a new one.
func (c *Counter) Count(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	c.evict(key, now)
	return len(c.entries[key])
}

// Reset clears all recorded occurrences for key.
func (c *Counter) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// evict removes timestamps outside the rolling window. Must be called
// with c.mu held.
func (c *Counter) evict(key string, now time.Time) {
	cutoff := now.Add(-c.window)
	times := c.entries[key]
	i := 0
	for i < len(times) && times[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		c.entries[key] = times[i:]
	}
	if len(c.entries[key]) == 0 {
		delete(c.entries, key)
	}
}
