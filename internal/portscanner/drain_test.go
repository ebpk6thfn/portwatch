package portscanner

import (
	"testing"
	"time"
)

func makeDrainEvent(port uint16, kind string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{Port: port, Protocol: "tcp"},
		Kind:  kind,
	}
}

func TestDrainer_PushBelowMax_NoFlush(t *testing.T) {
	d := NewDrainer(DrainPolicy{MaxBuffer: 4, MaxAge: time.Minute})
	out := d.Push(makeDrainEvent(80, "opened"))
	if out != nil {
		t.Fatalf("expected no flush, got %d events", len(out))
	}
	if d.Len() != 1 {
		t.Fatalf("expected len 1, got %d", d.Len())
	}
}

func TestDrainer_PushAtMax_Flushes(t *testing.T) {
	d := NewDrainer(DrainPolicy{MaxBuffer: 3, MaxAge: time.Minute})
	d.Push(makeDrainEvent(80, "opened"))
	d.Push(makeDrainEvent(443, "opened"))
	out := d.Push(makeDrainEvent(8080, "opened"))
	if len(out) != 3 {
		t.Fatalf("expected 3 flushed events, got %d", len(out))
	}
	if d.Len() != 0 {
		t.Fatalf("expected empty buffer after flush")
	}
}

func TestDrainer_Tick_NoFlush_WhenFresh(t *testing.T) {
	now := time.Now()
	d := NewDrainer(DrainPolicy{MaxBuffer: 10, MaxAge: time.Minute})
	d.now = func() time.Time { return now }
	d.Push(makeDrainEvent(22, "opened"))
	out := d.Tick()
	if out != nil {
		t.Fatalf("expected no flush on fresh entry")
	}
}

func TestDrainer_Tick_FlushesWhenExpired(t *testing.T) {
	now := time.Now()
	d := NewDrainer(DrainPolicy{MaxBuffer: 10, MaxAge: 5 * time.Second})
	d.now = func() time.Time { return now }
	d.Push(makeDrainEvent(22, "opened"))
	d.now = func() time.Time { return now.Add(6 * time.Second) }
	out := d.Tick()
	if len(out) != 1 {
		t.Fatalf("expected 1 flushed event, got %d", len(out))
	}
	if d.Len() != 0 {
		t.Fatalf("expected empty buffer after age flush")
	}
}

func TestDrainer_Tick_EmptyBuffer_NoOp(t *testing.T) {
	d := NewDrainer(DefaultDrainPolicy())
	out := d.Tick()
	if out != nil {
		t.Fatalf("expected nil on empty tick")
	}
}

func TestDrainer_DefaultPolicy_Applied(t *testing.T) {
	d := NewDrainer(DrainPolicy{})
	if d.policy.MaxBuffer != 64 {
		t.Fatalf("expected default MaxBuffer 64, got %d", d.policy.MaxBuffer)
	}
	if d.policy.MaxAge != 10*time.Second {
		t.Fatalf("expected default MaxAge 10s, got %v", d.policy.MaxAge)
	}
}
