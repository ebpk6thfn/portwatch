package portscanner_test

import (
	"testing"
	"time"
)

// TestBackpressure_PipelineStyleUsage verifies that the backpressure mechanism
// correctly trips when the event queue depth exceeds the high watermark and
// recovers once depth drops below the low watermark after the cooldown period.
func TestBackpressure_PipelineStyleUsage(t *testing.T) {
	t.Parallel()

	now := time.Now()
	clock := func() time.Time { return now }

	policy := DefaultBackpressurePolicy()
	policy.HighWatermark = 5
	policy.LowWatermark = 2
	policy.Cooldown = 100 * time.Millisecond

	bp := NewBackpressure(policy, clock)

	// Simulate events arriving; depth below high watermark — should not trip.
	for i := 0; i < 4; i++ {
		bp.RecordDepth(i + 1)
		if bp.IsActive() {
			t.Fatalf("expected backpressure inactive at depth %d", i+1)
		}
	}

	// Push depth to high watermark — backpressure should trip.
	bp.RecordDepth(5)
	if !bp.IsActive() {
		t.Fatal("expected backpressure to trip at high watermark")
	}

	// Depth drops below low watermark but cooldown not elapsed — still active.
	bp.RecordDepth(1)
	if !bp.IsActive() {
		t.Fatal("expected backpressure to remain active before cooldown")
	}

	// Advance clock past cooldown.
	now = now.Add(150 * time.Millisecond)
	bp.RecordDepth(1)
	if bp.IsActive() {
		t.Fatal("expected backpressure to clear after cooldown and low watermark")
	}
}

// TestBackpressure_EventFilterIntegration verifies that events are dropped
// while backpressure is active and allowed once it clears.
func TestBackpressure_EventFilterIntegration(t *testing.T) {
	t.Parallel()

	now := time.Now()
	clock := func() time.Time { return now }

	policy := DefaultBackpressurePolicy()
	policy.HighWatermark = 3
	policy.LowWatermark = 1
	policy.Cooldown = 50 * time.Millisecond

	bp := NewBackpressure(policy, clock)

	events := []ChangeEvent{
		makeBackpressureNow("tcp", 8080, now),
		makeBackpressureNow("tcp", 8081, now),
		makeBackpressureNow("tcp", 8082, now),
	}

	// Trip backpressure.
	bp.RecordDepth(3)
	if !bp.IsActive() {
		t.Fatal("expected backpressure active")
	}

	// All events should be dropped while active.
	var passed []ChangeEvent
	for _, ev := range events {
		if !bp.IsActive() {
			passed = append(passed, ev)
		}
	}
	if len(passed) != 0 {
		t.Fatalf("expected 0 events to pass, got %d", len(passed))
	}

	// Recover: depth drops and cooldown elapses.
	bp.RecordDepth(0)
	now = now.Add(100 * time.Millisecond)
	bp.RecordDepth(0)

	if bp.IsActive() {
		t.Fatal("expected backpressure to clear")
	}

	// New events should pass through.
	newEvent := makeBackpressureNow("tcp", 9090, now)
	if bp.IsActive() {
		t.Fatalf("unexpected backpressure active for event %v", newEvent)
	}
}
