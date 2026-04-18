package portscanner

import (
	"sync/atomic"
	"time"
)

// RateLimitMetrics tracks rate limiter activity for observability.
type RateLimitMetrics struct {
	allowed  atomic.Int64
	suppressed atomic.Int64
	lastReset time.Time
}

// NewRateLimitMetrics creates a fresh RateLimitMetrics.
func NewRateLimitMetrics() *RateLimitMetrics {
	return &RateLimitMetrics{lastReset: time.Now()}
}

// RecordAllowed increments the allowed counter.
func (m *RateLimitMetrics) RecordAllowed() {
	m.allowed.Add(1)
}

// RecordSuppressed increments the suppressed counter.
func (m *RateLimitMetrics) RecordSuppressed() {
	m.suppressed.Add(1)
}

// Allowed returns the total allowed count.
func (m *RateLimitMetrics) Allowed() int64 {
	return m.allowed.Load()
}

// Suppressed returns the total suppressed count.
func (m *RateLimitMetrics) Suppressed() int64 {
	return m.suppressed.Load()
}

// Reset zeroes all counters and records the reset time.
func (m *RateLimitMetrics) Reset() {
	m.allowed.Store(0)
	m.suppressed.Store(0)
	m.lastReset = time.Now()
}

// LastReset returns when the metrics were last reset.
func (m *RateLimitMetrics) LastReset() time.Time {
	return m.lastReset
}

// Summary returns a snapshot struct for logging or reporting.
type RateLimitSummary struct {
	Allowed    int64
	Suppressed int64
	LastReset  time.Time
}

// Summary returns a point-in-time snapshot.
func (m *RateLimitMetrics) Summary() RateLimitSummary {
	return RateLimitSummary{
		Allowed:    m.allowed.Load(),
		Suppressed: m.suppressed.Load(),
		LastReset:  m.lastReset,
	}
}
