package portscanner

import (
	"sync"
	"time"
)

// BurstDetector flags when the number of events within a rolling window
// exceeds a configured threshold.
type BurstDetector struct {
	mu        sync.Mutex
	threshold int
	window    time.Duration
	times     []time.Time
	now       func() time.Time
}

// NewBurstDetector creates a BurstDetector that fires when more than
// threshold events occur within window.
func NewBurstDetector(threshold int, window time.Duration) *BurstDetector {
	return &BurstDetector{
		threshold: threshold,
		window:    window,
		now:       time.Now,
	}
}

// Record adds an event timestamp and returns true if the burst threshold
// has been exceeded within the current window.
func (b *BurstDetector) Record() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	cutoff := now.Add(-b.window)

	// evict old entries
	filtered := b.times[:0]
	for _, t := range b.times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	b.times = append(filtered, now)

	return len(b.times) > b.threshold
}

// Count returns the number of events currently within the window.
func (b *BurstDetector) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	cutoff := now.Add(-b.window)
	count := 0
	for _, t := range b.times {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events.
func (b *BurstDetector) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.times = b.times[:0]
}
