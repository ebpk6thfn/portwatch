package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// RateLimitAuditEntry records a single rate-limit decision.
type RateLimitAuditEntry struct {
	Key       string
	Allowed   bool
	Reason    string
	Timestamp time.Time
}

// String returns a human-readable representation of the audit entry.
func (e RateLimitAuditEntry) String() string {
	state := "ALLOWED"
	if !e.Allowed {
		state = "SUPPRESSED"
	}
	return fmt.Sprintf("%s [%s] key=%s reason=%s",
		e.Timestamp.Format(time.RFC3339), state, e.Key, e.Reason)
}

// RateLimitAuditLog retains a bounded history of rate-limit decisions
// for debugging and observability.
type RateLimitAuditLog struct {
	mu      sync.Mutex
	entries []RateLimitAuditEntry
	maxSize int
}

// NewRateLimitAuditLog creates an audit log with the given capacity.
// If maxSize <= 0 it defaults to 256.
func NewRateLimitAuditLog(maxSize int) *RateLimitAuditLog {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &RateLimitAuditLog{maxSize: maxSize}
}

// Record appends a decision to the log, evicting the oldest entry when full.
func (a *RateLimitAuditLog) Record(key string, allowed bool, reason string, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	entry := RateLimitAuditEntry{
		Key:       key,
		Allowed:   allowed,
		Reason:    reason,
		Timestamp: now,
	}
	if len(a.entries) >= a.maxSize {
		a.entries = a.entries[1:]
	}
	a.entries = append(a.entries, entry)
}

// All returns a snapshot of all recorded entries in insertion order.
func (a *RateLimitAuditLog) All() []RateLimitAuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]RateLimitAuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}

// Len returns the current number of entries.
func (a *RateLimitAuditLog) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.entries)
}

// Clear removes all entries from the log.
func (a *RateLimitAuditLog) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = a.entries[:0]
}
