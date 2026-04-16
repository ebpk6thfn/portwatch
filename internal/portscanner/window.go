package portscanner

import (
	"sync"
	"time"
)

// WindowCounter counts events within a sliding time window per key.
type WindowCounter struct {
	mu       sync.Mutex
	window   time.Duration
	buckets  map[string][]time.Time
	nowFn    func() time.Time
}

// NewWindowCounter creates a WindowCounter with the given sliding window duration.
func NewWindowCounter(window time.Duration) *WindowCounter {
	return &WindowCounter{
		window:  window,
		buckets: make(map[string][]time.Time),
		nowFn:   time.Now,
	}
}

// Add records an event for key and returns the total count within the window.
func (w *WindowCounter) Add(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.nowFn()
	w.buckets[key] = append(w.prune(key, now), now)
	return len(w.buckets[key])
}

// Count returns the current count for key without adding a new event.
func (w *WindowCounter) Count(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.prune(key, w.nowFn()))
}

// Reset clears all recorded events for key.
func (w *WindowCounter) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.buckets, key)
}

// prune removes timestamps outside the window and updates the bucket.
func (w *WindowCounter) prune(key string, now time.Time) []time.Time {
	cutoff := now.Add(-w.window)
	ts := w.buckets[key]
	i := 0
	for i < len(ts) && ts[i].Before(cutoff) {
		i++
	}
	pruned := ts[i:]
	w.buckets[key] = pruned
	return pruned
}
