package portscanner

import (
	"sync"
	"time"
)

// DecayCounter tracks event counts that decay (expire) over a sliding window.
// Unlike WindowCounter, it returns a weighted score that decreases as events age.
type DecayCounter struct {
	mu       sync.Mutex
	events   []time.Time
	window   time.Duration
	now      func() time.Time
}

// NewDecayCounter creates a DecayCounter with the given sliding window.
func NewDecayCounter(window time.Duration) *DecayCounter {
	return &DecayCounter{
		window: window,
		now:    time.Now,
	}
}

// Add records a new event at the current time.
func (d *DecayCounter) Add() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events = append(d.events, d.now())
	d.evict()
}

// Score returns a float in [0,1] representing how "recent" the burst is.
// Events closer to now contribute more weight than older ones.
func (d *DecayCounter) Score() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.evict()
	if len(d.events) == 0 {
		return 0
	}
	now := d.now()
	var score float64
	for _, t := range d.events {
		age := now.Sub(t)
		// Linear decay: weight = 1 - (age / window)
		weight := 1.0 - float64(age)/float64(d.window)
		if weight > 0 {
			score += weight
		}
	}
	return score
}

// Count returns the raw number of unexpired events.
func (d *DecayCounter) Count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.evict()
	return len(d.events)
}

// evict removes events outside the window. Must be called with lock held.
func (d *DecayCounter) evict() {
	cutoff := d.now().Add(-d.window)
	i := 0
	for i < len(d.events) && d.events[i].Before(cutoff) {
		i++
	}
	d.events = d.events[i:]
}
