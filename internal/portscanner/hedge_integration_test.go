package portscanner

import (
	"testing"
	"time"
)

// TestHedge_PipelineStyleUsage simulates a realistic pipeline scenario where
// short-lived port flaps are suppressed and stable opens are eventually emitted.
func TestHedge_PipelineStyleUsage(t *testing.T) {
	base := time.Now()
	var current time.Time = base
	clk := func() time.Time { return current }

	h := newHedgeWithClock(HedgePolicy{Window: 2 * time.Second, MaxPending: 64}, clk)

	// Port 8080 opens then closes quickly — should be cancelled.
	h.Hold(makeHedgeEvent(8080, "tcp", EventOpened))
	h.Hold(makeHedgeEvent(8080, "tcp", EventClosed))

	// Port 9090 opens and stays open past the window.
	h.Hold(makeHedgeEvent(9090, "tcp", EventOpened))

	// Before window: no events emitted.
	current = base.Add(1 * time.Second)
	events := h.Flush()
	if len(events) != 0 {
		t.Fatalf("expected no events before window, got %d", len(events))
	}

	// After window: only the stable open for 9090 should appear.
	current = base.Add(3 * time.Second)
	events = h.Flush()
	if len(events) != 1 {
		t.Fatalf("expected 1 event after window, got %d", len(events))
	}
	if events[0].Entry.Port != 9090 {
		t.Errorf("expected port 9090, got %d", events[0].Entry.Port)
	}
	if events[0].Type != EventOpened {
		t.Errorf("expected EventOpened, got %s", events[0].Type)
	}
}
