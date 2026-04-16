package portscanner

import (
	"sync"
	"time"
)

// TrendDirection indicates whether port activity is increasing or decreasing.
type TrendDirection string

const (
	TrendUp     TrendDirection = "up"
	TrendDown   TrendDirection = "down"
	TrendStable TrendDirection = "stable"
)

// TrendPoint is a single observation in the trend window.
type TrendPoint struct {
	At    time.Time
	Count int
}

// TrendTracker tracks event frequency over a sliding window to compute trends.
type TrendTracker struct {
	mu     sync.Mutex
	window time.Duration
	points []TrendPoint
}

// NewTrendTracker creates a TrendTracker with the given sliding window.
func NewTrendTracker(window time.Duration) *TrendTracker {
	return &TrendTracker{window: window}
}

// Record adds a new observation at the given time with the given event count.
func (t *TrendTracker) Record(at time.Time, count int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.evict(at)
	t.points = append(t.points, TrendPoint{At: at, Count: count})
}

// Trend returns the current trend direction based on recorded points.
func (t *TrendTracker) Trend(now time.Time) TrendDirection {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.evict(now)
	if len(t.points) < 2 {
		return TrendStable
	}
	first := t.points[0].Count
	last := t.points[len(t.points)-1].Count
	switch {
	case last > first:
		return TrendUp
	case last < first:
		return TrendDown
	default:
		return TrendStable
	}
}

// Points returns a copy of the current window's points.
func (t *TrendTracker) Points(now time.Time) []TrendPoint {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.evict(now)
	out := make([]TrendPoint, len(t.points))
	copy(out, t.points)
	return out
}

func (t *TrendTracker) evict(now time.Time) {
	cutoff := now.Add(-t.window)
	i := 0
	for i < len(t.points) && t.points[i].At.Before(cutoff) {
		i++
	}
	t.points = t.points[i:]
}
