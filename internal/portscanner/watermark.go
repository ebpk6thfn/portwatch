package portscanner

import (
	"sync"
	"time"
)

// WatermarkPolicy configures high/low watermark thresholds for event volume.
type WatermarkPolicy struct {
	HighMark  int
	LowMark   int
	Window    time.Duration
	Cooldown  time.Duration
}

// DefaultWatermarkPolicy returns sensible defaults.
func DefaultWatermarkPolicy() WatermarkPolicy {
	return WatermarkPolicy{
		HighMark: 50,
		LowMark:  20,
		Window:   time.Minute,
		Cooldown: 2 * time.Minute,
	}
}

// WatermarkState tracks whether the watermark is currently breached.
type WatermarkState int

const (
	WatermarkNormal  WatermarkState = iota
	WatermarkBreached
)

// Watermark tracks event volume against high/low thresholds within a sliding window.
type Watermark struct {
	mu       sync.Mutex
	policy   WatermarkPolicy
	events   []time.Time
	state    WatermarkState
	cooledAt time.Time
	now      func() time.Time
}

// NewWatermark creates a Watermark with the given policy.
func NewWatermark(p WatermarkPolicy, now func() time.Time) *Watermark {
	if now == nil {
		now = time.Now
	}
	return &Watermark{policy: p, now: now}
}

// Record adds an event timestamp and returns the current WatermarkState.
func (w *Watermark) Record() WatermarkState {
	w.mu.Lock()
	defer w.mu.Unlock()

	t := w.now()
	w.evict(t)
	w.events = append(w.events, t)

	switch w.state {
	case WatermarkNormal:
		if len(w.events) >= w.policy.HighMark {
			w.state = WatermarkBreached
			w.cooledAt = time.Time{}
		}
	case WatermarkBreached:
		if len(w.events) <= w.policy.LowMark {
			if w.cooledAt.IsZero() {
				w.cooledAt = t
			} else if t.Sub(w.cooledAt) >= w.policy.Cooldown {
				w.state = WatermarkNormal
				w.cooledAt = time.Time{}
			}
		} else {
			w.cooledAt = time.Time{}
		}
	}
	return w.state
}

// State returns the current watermark state without recording an event.
func (w *Watermark) State() WatermarkState {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state
}

// Depth returns the number of events currently within the window.
func (w *Watermark) Depth() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(w.now())
	return len(w.events)
}

func (w *Watermark) evict(now time.Time) {
	cutoff := now.Add(-w.policy.Window)
	i := 0
	for i < len(w.events) && w.events[i].Before(cutoff) {
		i++
	}
	w.events = w.events[i:]
}
