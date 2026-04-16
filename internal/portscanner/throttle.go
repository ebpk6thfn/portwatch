package portscanner

import (
	"sync"
	"time"
)

// ThrottleConfig controls how the Throttle behaves.
type ThrottleConfig struct {
	// MaxPerInterval is the maximum number of events allowed per interval.
	MaxPerInterval int
	// Interval is the rolling window duration.
	Interval time.Duration
}

// Throttle limits the total number of ChangeEvents emitted within a rolling
// time window, regardless of which port or key they belong to.
type Throttle struct {
	mu       sync.Mutex
	cfg      ThrottleConfig
	timings  []time.Time
	nowFn    func() time.Time
}

// NewThrottle creates a Throttle with the given configuration.
func NewThrottle(cfg ThrottleConfig) *Throttle {
	return &Throttle{
		cfg:   cfg,
		nowFn: time.Now,
	}
}

// Allow returns true if the event should be forwarded, false if it should be
// dropped because the global rate limit has been reached.
func (t *Throttle) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.nowFn()
	cutoff := now.Add(-t.cfg.Interval)

	// Evict timestamps outside the rolling window.
	valid := t.timings[:0]
	for _, ts := range t.timings {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	t.timings = valid

	if t.cfg.MaxPerInterval > 0 && len(t.timings) >= t.cfg.MaxPerInterval {
		return false
	}

	t.timings = append(t.timings, now)
	return true
}

// Remaining returns how many more events can pass in the current window.
func (t *Throttle) Remaining() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cfg.MaxPerInterval <= 0 {
		return -1 // unlimited
	}
	now := t.nowFn()
	cutoff := now.Add(-t.cfg.Interval)
	count := 0
	for _, ts := range t.timings {
		if ts.After(cutoff) {
			count++
		}
	}
	r := t.cfg.MaxPerInterval - count
	if r < 0 {
		return 0
	}
	return r
}

// Filter applies the throttle to a slice of ChangeEvents, returning only those
// that are allowed through.
func (t *Throttle) Filter(events []ChangeEvent) []ChangeEvent {
	out := make([]ChangeEvent, 0, len(events))
	for _, e := range events {
		if t.Allow() {
			out = append(out, e)
		}
	}
	return out
}
