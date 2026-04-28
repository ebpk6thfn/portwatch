package portscanner

import (
	"testing"
	"time"
)

func makeHedgeNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func makeHedgeEvent(port uint16, proto, evType string) ChangeEvent {
	return ChangeEvent{
		Type: evType,
		Entry: Entry{
			Port:     port,
			Protocol: proto,
		},
	}
}

func TestHedge_HoldNewEvent_ReturnsTrue(t *testing.T) {
	base := time.Now()
	clk := makeHedgeNow(base)
	h := newHedgeWithClock(DefaultHedgePolicy(), clk)

	ev := makeHedgeEvent(8080, "tcp", EventOpened)
	if !h.Hold(ev) {
		t.Fatal("expected Hold to return true for new event")
	}
	if h.Len() != 1 {
		t.Fatalf("expected 1 pending, got %d", h.Len())
	}
}

func TestHedge_OpposingEvent_CancelsBoth(t *testing.T) {
	base := time.Now()
	clk := makeHedgeNow(base)
	h := newHedgeWithClock(DefaultHedgePolicy(), clk)

	open := makeHedgeEvent(8080, "tcp", EventOpened)
	close := makeHedgeEvent(8080, "tcp", EventClosed)

	h.Hold(open)
	result := h.Hold(close)
	if result {
		t.Fatal("expected Hold to return false when opposing event cancels")
	}
	if h.Len() != 0 {
		t.Fatalf("expected 0 pending after cancellation, got %d", h.Len())
	}
}

func TestHedge_Flush_BeforeWindow_ReturnsNothing(t *testing.T) {
	base := time.Now()
	clk := makeHedgeNow(base)
	h := newHedgeWithClock(HedgePolicy{Window: 5 * time.Second, MaxPending: 64}, clk)

	h.Hold(makeHedgeEvent(9000, "tcp", EventOpened))

	events := h.Flush()
	if len(events) != 0 {
		t.Fatalf("expected 0 events before window, got %d", len(events))
	}
	if h.Len() != 1 {
		t.Fatalf("expected 1 still pending, got %d", h.Len())
	}
}

func TestHedge_Flush_AfterWindow_EmitsEvent(t *testing.T) {
	base := time.Now()
	var current time.Time = base
	clk := func() time.Time { return current }
	h := newHedgeWithClock(HedgePolicy{Window: 2 * time.Second, MaxPending: 64}, clk)

	ev := makeHedgeEvent(443, "tcp", EventOpened)
	h.Hold(ev)

	current = base.Add(3 * time.Second)
	events := h.Flush()
	if len(events) != 1 {
		t.Fatalf("expected 1 event after window, got %d", len(events))
	}
	if events[0].Entry.Port != 443 {
		t.Errorf("unexpected port: %d", events[0].Entry.Port)
	}
	if h.Len() != 0 {
		t.Fatalf("expected 0 pending after flush, got %d", h.Len())
	}
}

func TestHedge_MaxPending_EvictsOldest(t *testing.T) {
	base := time.Now()
	var current time.Time = base
	clk := func() time.Time { return current }
	h := newHedgeWithClock(HedgePolicy{Window: 10 * time.Second, MaxPending: 2}, clk)

	h.Hold(makeHedgeEvent(1000, "tcp", EventOpened))
	current = base.Add(1 * time.Second)
	h.Hold(makeHedgeEvent(1001, "tcp", EventOpened))
	current = base.Add(2 * time.Second)
	h.Hold(makeHedgeEvent(1002, "tcp", EventOpened))

	if h.Len() != 2 {
		t.Fatalf("expected 2 pending after eviction, got %d", h.Len())
	}
}

func TestHedge_SameDirectionSameKey_NoCancel(t *testing.T) {
	base := time.Now()
	clk := makeHedgeNow(base)
	h := newHedgeWithClock(DefaultHedgePolicy(), clk)

	h.Hold(makeHedgeEvent(7777, "udp", EventOpened))
	result := h.Hold(makeHedgeEvent(7777, "udp", EventOpened))
	if !result {
		t.Fatal("expected Hold to return true for same-direction duplicate")
	}
	if h.Len() != 1 {
		t.Fatalf("expected 1 pending, got %d", h.Len())
	}
}
