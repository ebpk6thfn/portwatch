package portscanner

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks runtime counters for the portscanner pipeline.
type Metrics struct {
	mu sync.RWMutex

	ScansTotal     uint64
	EventsEmitted  uint64
	EventsDropped  uint64
	LastScanAt     time.Time
	LastScanDur    time.Duration
}

// global singleton for package-level access.
var globalMetrics = &Metrics{}

// GlobalMetrics returns the package-level Metrics instance.
func GlobalMetrics() *Metrics {
	return globalMetrics
}

// RecordScan increments the scan counter and records timing information.
func (m *Metrics) RecordScan(dur time.Duration, at time.Time) {
	atomic.AddUint64(&m.ScansTotal, 1)
	m.mu.Lock()
	m.LastScanAt = at
	m.LastScanDur = dur
	m.mu.Unlock()
}

// RecordEmitted increments the emitted event counter by n.
func (m *Metrics) RecordEmitted(n int) {
	atomic.AddUint64(&m.EventsEmitted, uint64(n))
}

// RecordDropped increments the dropped event counter by n.
func (m *Metrics) RecordDropped(n int) {
	atomic.AddUint64(&m.EventsDropped, uint64(n))
}

// Snapshot returns a point-in-time copy of the current metrics.
func (m *Metrics) Snapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Metrics{
		ScansTotal:    atomic.LoadUint64(&m.ScansTotal),
		EventsEmitted: atomic.LoadUint64(&m.EventsEmitted),
		EventsDropped: atomic.LoadUint64(&m.EventsDropped),
		LastScanAt:    m.LastScanAt,
		LastScanDur:   m.LastScanDur,
	}
}

// Reset zeroes all counters. Primarily useful in tests.
func (m *Metrics) Reset() {
	atomic.StoreUint64(&m.ScansTotal, 0)
	atomic.StoreUint64(&m.EventsEmitted, 0)
	atomic.StoreUint64(&m.EventsDropped, 0)
	m.mu.Lock()
	m.LastScanAt = time.Time{}
	m.LastScanDur = 0
	m.mu.Unlock()
}
