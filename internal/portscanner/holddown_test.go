package portscanner

import (
	"net"
	"testing"
	"time"
)

func makeHolddownEvent(t EventType, port uint16, proto string) ChangeEvent {
	return ChangeEvent{
		Type: t,
		Entry: Entry{
			LocalIP:  net.ParseIP("127.0.0.1"),
			Port:     port,
			Protocol: proto,
		},
	}
}

func TestHolddown_OpenedPassesImmediately(t *testing.T) {
	h := NewHolddown(DefaultHolddownPolicy())
	events := []ChangeEvent{makeHolddownEvent(EventOpened, 8080, "tcp")}
	out := h.Evaluate(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
}

func TestHolddown_ClosedWithinWindow_Suppressed(t *testing.T) {
	now := time.Now()
	h := NewHolddown(HolddownPolicy{Duration: 10 * time.Second})
	h.nowFn = func() time.Time { return now }

	events := []ChangeEvent{makeHolddownEvent(EventClosed, 9000, "tcp")}
	out := h.Evaluate(events)
	if len(out) != 0 {
		t.Fatalf("expected 0 events within hold-down, got %d", len(out))
	}
	if h.PendingCount() != 1 {
		t.Fatalf("expected 1 pending entry, got %d", h.PendingCount())
	}
}

func TestHolddown_ClosedAfterWindow_Emitted(t *testing.T) {
	now := time.Now()
	h := NewHolddown(HolddownPolicy{Duration: 5 * time.Second})
	h.nowFn = func() time.Time { return now }

	events := []ChangeEvent{makeHolddownEvent(EventClosed, 9000, "tcp")}
	h.Evaluate(events) // first pass — starts timer

	// Advance past the hold-down window.
	h.nowFn = func() time.Time { return now.Add(6 * time.Second) }
	out := h.Evaluate(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event after hold-down, got %d", len(out))
	}
	if h.PendingCount() != 0 {
		t.Fatalf("expected pending count 0 after emit, got %d", h.PendingCount())
	}
}

func TestHolddown_ReopenCancelsPending(t *testing.T) {
	now := time.Now()
	h := NewHolddown(HolddownPolicy{Duration: 10 * time.Second})
	h.nowFn = func() time.Time { return now }

	closed := []ChangeEvent{makeHolddownEvent(EventClosed, 443, "tcp")}
	h.Evaluate(closed)
	if h.PendingCount() != 1 {
		t.Fatal("expected pending after close")
	}

	opened := []ChangeEvent{makeHolddownEvent(EventOpened, 443, "tcp")}
	out := h.Evaluate(opened)
	if len(out) != 1 {
		t.Fatalf("expected opened event, got %d", len(out))
	}
	if h.PendingCount() != 0 {
		t.Fatalf("expected pending cleared, got %d", h.PendingCount())
	}
}

func TestHolddown_ZeroDuration_PassesAll(t *testing.T) {
	h := NewHolddown(HolddownPolicy{Duration: 0})
	events := []ChangeEvent{
		makeHolddownEvent(EventClosed, 80, "tcp"),
		makeHolddownEvent(EventClosed, 443, "tcp"),
	}
	out := h.Evaluate(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events with zero duration, got %d", len(out))
	}
}

func TestHolddown_Flush_EmitsPending(t *testing.T) {
	now := time.Now()
	h := NewHolddown(HolddownPolicy{Duration: 30 * time.Second})
	h.nowFn = func() time.Time { return now }

	events := []ChangeEvent{makeHolddownEvent(EventClosed, 8080, "tcp")}
	h.Evaluate(events)

	flushed := h.Flush(func(key string) ChangeEvent {
		return makeHolddownEvent(EventClosed, 8080, "tcp")
	})
	if len(flushed) != 1 {
		t.Fatalf("expected 1 flushed event, got %d", len(flushed))
	}
	if h.PendingCount() != 0 {
		t.Fatal("expected pending cleared after flush")
	}
}
