package portscanner

import (
	"testing"
	"time"
)

func makeThrottleEvent(port uint16, kind ChangeKind) ChangeEvent {
	return ChangeEvent{
		Kind:  kind,
		Entry: Entry{Port: port, Protocol: "tcp"},
	}
}

func TestThrottle_UnlimitedAlwaysAllows(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 0, Interval: time.Second})
	for i := 0; i < 100; i++ {
		if !th.Allow() {
			t.Fatalf("unlimited throttle should always allow, failed at iteration %d", i)
		}
	}
}

func TestThrottle_BlocksAfterMax(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 3, Interval: time.Minute})
	for i := 0; i < 3; i++ {
		if !th.Allow() {
			t.Fatalf("should allow event %d", i)
		}
	}
	if th.Allow() {
		t.Fatal("should have been blocked after max")
	}
}

func TestThrottle_ResetsAfterInterval(t *testing.T) {
	now := time.Now()
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 2, Interval: time.Second})
	th.nowFn = func() time.Time { return now }

	th.Allow()
	th.Allow()
	if th.Allow() {
		t.Fatal("should be blocked at limit")
	}

	// Advance time past the interval.
	th.nowFn = func() time.Time { return now.Add(2 * time.Second) }
	if !th.Allow() {
		t.Fatal("should be allowed after interval reset")
	}
}

func TestThrottle_Remaining_Decrements(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 5, Interval: time.Minute})
	if th.Remaining() != 5 {
		t.Fatalf("expected 5 remaining, got %d", th.Remaining())
	}
	th.Allow()
	if th.Remaining() != 4 {
		t.Fatalf("expected 4 remaining, got %d", th.Remaining())
	}
}

func TestThrottle_Remaining_UnlimitedIsNegativeOne(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 0, Interval: time.Minute})
	if th.Remaining() != -1 {
		t.Fatalf("expected -1 for unlimited, got %d", th.Remaining())
	}
}

func TestThrottle_Filter_DropsExcess(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 2, Interval: time.Minute})
	events := []ChangeEvent{
		makeThrottleEvent(80, PortOpened),
		makeThrottleEvent(443, PortOpened),
		makeThrottleEvent(8080, PortOpened),
	}
	out := th.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events after throttle, got %d", len(out))
	}
}

func TestThrottle_Filter_AllowsWhenUnderLimit(t *testing.T) {
	th := NewThrottle(ThrottleConfig{MaxPerInterval: 10, Interval: time.Minute})
	events := []ChangeEvent{
		makeThrottleEvent(80, PortOpened),
		makeThrottleEvent(443, PortClosed),
	}
	out := th.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out))
	}
}
