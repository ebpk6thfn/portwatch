package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// SpikePolicy configures spike detection behaviour.
type SpikePolicy struct {
	// Window is the duration over which events are counted.
	Window time.Duration
	// Threshold is the number of events in Window that triggers a spike.
	Threshold int
	// Cooldown suppresses repeated spike alerts for this duration.
	Cooldown time.Duration
}

// DefaultSpikePolicy returns a sensible default spike policy.
func DefaultSpikePolicy() SpikePolicy {
	return SpikePolicy{
		Window:    30 * time.Second,
		Threshold: 10,
		Cooldown:  2 * time.Minute,
	}
}

// SpikeAlert is emitted when a spike is detected.
type SpikeAlert struct {
	Count     int
	Window    time.Duration
	Threshold int
	DetectedAt time.Time
}

func (s SpikeAlert) String() string {
	return fmt.Sprintf("spike: %d events in %s (threshold %d) at %s",
		s.Count, s.Window, s.Threshold, s.DetectedAt.Format(time.RFC3339))
}

// SpikeDetector tracks event counts over a sliding window and fires
// a SpikeAlert when the count exceeds the configured threshold.
type SpikeDetector struct {
	mu       sync.Mutex
	policy   SpikePolicy
	timestamps []time.Time
	lastAlert  time.Time
	now        func() time.Time
}

// NewSpikeDetector creates a SpikeDetector with the given policy.
func NewSpikeDetector(policy SpikePolicy) *SpikeDetector {
	return &SpikeDetector{
		policy: policy,
		now:    time.Now,
	}
}

// Record adds an event timestamp and returns a SpikeAlert if the
// threshold is exceeded and the cooldown has elapsed, otherwise nil.
func (d *SpikeDetector) Record(events []ChangeEvent) *SpikeAlert {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	cutoff := now.Add(-d.policy.Window)

	// Append new events.
	for range events {
		d.timestamps = append(d.timestamps, now)
	}

	// Evict old timestamps.
	valid := d.timestamps[:0]
	for _, ts := range d.timestamps {
		if !ts.Before(cutoff) {
			valid = append(valid, ts)
		}
	}
	d.timestamps = valid

	count := len(d.timestamps)
	if count < d.policy.Threshold {
		return nil
	}

	// Respect cooldown.
	if !d.lastAlert.IsZero() && now.Sub(d.lastAlert) < d.policy.Cooldown {
		return nil
	}

	d.lastAlert = now
	return &SpikeAlert{
		Count:      count,
		Window:     d.policy.Window,
		Threshold:  d.policy.Threshold,
		DetectedAt: now,
	}
}

// Count returns the number of events currently within the window.
func (d *SpikeDetector) Count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.timestamps)
}
