package portscanner

import (
	"testing"
	"time"
)

func makeBackpressureNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestBackpressure_InitiallyInactive(t *testing.T) {
	now := makeBackpressureNow(time.Now())
	bp := NewBackpressure(DefaultBackpressurePolicy(), now)
	if bp.IsActive() {
		t.Fatal("expected backpressure to be inactive initially")
	}
}

func TestBackpressure_TripsAtHighWatermark(t *testing.T) {
	policy := BackpressurePolicy{HighWatermark: 3, LowWatermark: 1, CooldownPeriod: time.Second}
	now := makeBackpressureNow(time.Now())
	bp := NewBackpressure(policy, now)

	bp.Push()
	bp.Push()
	active := bp.Push() // 3rd push should trip
	if !active {
		t.Fatal("expected backpressure to be active at high watermark")
	}
}

func TestBackpressure_RemainsActiveUntilLowWatermarkAndCooldown(t *testing.T) {
	base := time.Now()
	current := base
	now := func() time.Time { return current }
	policy := BackpressurePolicy{HighWatermark: 2, LowWatermark: 1, CooldownPeriod: 2 * time.Second}
	bp := NewBackpressure(policy, now)

	bp.Push()
	bp.Push() // trips

	// Pop once — still above low watermark
	bp.Pop()
	if !bp.IsActive() {
		t.Fatal("expected still active above low watermark")
	}

	// Pop again — at low watermark but cooldown not elapsed
	bp.Pop()
	if !bp.IsActive() {
		t.Fatal("expected still active within cooldown period")
	}

	// Advance time past cooldown
	current = base.Add(3 * time.Second)
	bp.Pop() // depth goes to 0, cooldown elapsed
	if bp.IsActive() {
		t.Fatal("expected backpressure to release after low watermark and cooldown")
	}
}

func TestBackpressure_DepthTracking(t *testing.T) {
	now := makeBackpressureNow(time.Now())
	bp := NewBackpressure(DefaultBackpressurePolicy(), now)

	for i := 0; i < 5; i++ {
		bp.Push()
	}
	if bp.Depth() != 5 {
		t.Fatalf("expected depth 5, got %d", bp.Depth())
	}
	bp.Pop()
	bp.Pop()
	if bp.Depth() != 3 {
		t.Fatalf("expected depth 3, got %d", bp.Depth())
	}
}

func TestBackpressure_DepthNeverNegative(t *testing.T) {
	now := makeBackpressureNow(time.Now())
	bp := NewBackpressure(DefaultBackpressurePolicy(), now)
	bp.Pop()
	bp.Pop()
	if bp.Depth() != 0 {
		t.Fatalf("expected depth to floor at 0, got %d", bp.Depth())
	}
}
