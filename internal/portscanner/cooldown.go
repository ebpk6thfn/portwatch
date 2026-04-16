package portscanner

import (
	"sync"
	"time"
)

// Cooldown tracks per-key cooldown periods and reports whether an event
// should be allowed through or held back until the cooldown expires.
type Cooldown struct {
	mu       sync.Mutex
	period   time.Duration
	expiry   map[string]time.Time
	nowFn    func() time.Time
}

// NewCooldown creates a Cooldown with the given suppression period.
func NewCooldown(period time.Duration) *Cooldown {
	return &Cooldown{
		period: period,
		expiry: make(map[string]time.Time),
		nowFn:  time.Now,
	}
}

// Allow returns true if the key is not currently in cooldown and records
// a new cooldown window starting now. Returns false if suppressed.
func (c *Cooldown) Allow(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.nowFn()
	if exp, ok := c.expiry[key]; ok && now.Before(exp) {
		return false
	}
	c.expiry[key] = now.Add(c.period)
	return true
}

// Reset removes the cooldown entry for the given key immediately.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.expiry, key)
}

// Flush removes all expired entries to prevent unbounded growth.
func (c *Cooldown) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.nowFn()
	for k, exp := range c.expiry {
		if now.After(exp) {
			delete(c.expiry, k)
		}
	}
}

// Len returns the number of tracked keys.
func (c *Cooldown) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.expiry)
}
