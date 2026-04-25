package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// VelocityPolicy controls how velocity is calculated and reported.
type VelocityPolicy struct {
	Window   time.Duration
	MaxItems int
}

// DefaultVelocityPolicy returns a sensible default policy.
func DefaultVelocityPolicy() VelocityPolicy {
	return VelocityPolicy{
		Window:   time.Minute,
		MaxItems: 1000,
	}
}

type velocityEntry struct {
	at    time.Time
	count int
}

// Velocity tracks the rate of change events per key over a sliding window.
type Velocity struct {
	mu     sync.Mutex
	policy VelocityPolicy
	buckets map[string][]velocityEntry
	now    func() time.Time
}

// NewVelocity creates a new Velocity tracker.
func NewVelocity(policy VelocityPolicy, now func() time.Time) *Velocity {
	if now == nil {
		now = time.Now
	}
	return &Velocity{
		policy:  policy,
		buckets: make(map[string][]velocityEntry),
		now:     now,
	}
}

// Record adds an event for the given key and returns the current rate (events/min).
func (v *Velocity) Record(key string) float64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := v.now()
	v.evict(key, now)

	v.buckets[key] = append(v.buckets[key], velocityEntry{at: now, count: 1})
	if len(v.buckets[key]) > v.policy.MaxItems {
		v.buckets[key] = v.buckets[key][1:]
	}

	return v.rate(key, now)
}

// Rate returns the current event rate (events/min) for a key without recording.
func (v *Velocity) Rate(key string) float64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.evict(key, v.now())
	return v.rate(key, v.now())
}

func (v *Velocity) rate(key string, now time.Time) float64 {
	entries := v.buckets[key]
	if len(entries) == 0 {
		return 0
	}
	total := 0
	for _, e := range entries {
		total += e.count
	}
	windowSecs := v.policy.Window.Seconds()
	if windowSecs == 0 {
		return 0
	}
	return float64(total) / (windowSecs / 60.0)
}

func (v *Velocity) evict(key string, now time.Time) {
	cutoff := now.Add(-v.policy.Window)
	entries := v.buckets[key]
	i := 0
	for i < len(entries) && entries[i].at.Before(cutoff) {
		i++
	}
	v.buckets[key] = entries[i:]
}

// String returns a human-readable summary of the velocity for a key.
func (v *Velocity) String(key string) string {
	return fmt.Sprintf("velocity[%s]=%.2f events/min", key, v.Rate(key))
}
