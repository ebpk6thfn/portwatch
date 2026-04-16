package portscanner

import (
	"testing"
	"time"
)

// TestThrottleWithMetrics_DroppedCountedCorrectly verifies that when a Throttle
// drops events, the caller can correctly account for them via Metrics.
func TestThrottleWithMetrics_DroppedCountedCorrectly(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 2, Interval: time.Minute})
	m := freshMetrics()

	events := []ChangeEvent{
		makeThrottleEvent(80, PortOpened),
		makeThrottleEvent(443, PortOpened),
		makeThrottleEvent(8080, PortOpened),
		makeThrottleEvent(9090, PortOpened),
	}

	allowed := th.Filter(events)
	dropped := len(events) - len(allowed)

	m.RecordEmitted(len(allowed))
	m.RecordDropped(dropped)

	snap := m.Snapshot()
	if snap.EventsEmitted != 2 {
		t.Fatalf("expected 2 emitted, got %d", snap.EventsEmitted)
	}
	if snap.EventsDropped != 2 {
		t.Fatalf("expected 2 dropped, got %d", snap.EventsDropped)
	}
}

// TestThrottleWithMetrics_ScanRecordedWithDuration checks that scan timing
// integrates cleanly with a throttle-filtered pipeline step.
func TestThrottleWithMetrics_ScanRecordedWithDuration(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 10, Interval: time.Minute})
	m := freshMetrics()

	start := time.Now()
	events := []ChangeEvent{
		makeThrottleEvent(22, PortOpened),
		makeThrottleEvent(3306, PortClosed),
	}
	allowed := th.Filter(events)
	dur := time.Since(start)

	m.RecordScan(dur, start)
	m.RecordEmitted(len(allowed))

	snap := m.Snapshot()
	if snap.ScansTotal != 1 {
		t.Fatalf("expected 1 scan recorded, got %d", snap.ScansTotal)
	}
	if snap.EventsEmitted != 2 {
		t.Fatalf("expected 2 emitted, got %d", snap.EventsEmitted)
	}
	if snap.EventsDropped != 0 {
		t.Fatalf("expected 0 dropped, got %d", snap.EventsDropped)
	}
}
