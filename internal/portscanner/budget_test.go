package portscanner

import (
	"testing"
)

func makeBudgetEvent(port uint16) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{LocalPort: port, Protocol: "tcp"},
		Kind:  ChangeOpened,
	}
}

func TestBudget_UnlimitedAlwaysAllows(t *testing.T) {
	b := NewBudget(0)
	for i := 0; i < 1000; i++ {
		if !b.Allow() {
			t.Fatalf("unlimited budget denied at iteration %d", i)
		}
	}
}

func TestBudget_ExhaustsAtMax(t *testing.T) {
	b := NewBudget(3)
	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Fatalf("budget should allow up to max, denied at %d", i)
		}
	}
	if b.Allow() {
		t.Fatal("budget should be exhausted after max calls")
	}
}

func TestBudget_ResetRestoresFull(t *testing.T) {
	b := NewBudget(2)
	b.Allow()
	b.Allow()
	b.Reset()
	if !b.Allow() {
		t.Fatal("budget should allow after reset")
	}
}

func TestBudget_RemainingDecrementsCorrectly(t *testing.T) {
	b := NewBudget(5)
	if b.Remaining() != 5 {
		t.Fatalf("expected 5 remaining, got %d", b.Remaining())
	}
	b.Allow()
	if b.Remaining() != 4 {
		t.Fatalf("expected 4 remaining, got %d", b.Remaining())
	}
}

func TestBudget_Remaining_Unlimited(t *testing.T) {
	b := NewBudget(0)
	if b.Remaining() != -1 {
		t.Fatalf("expected -1 for unlimited, got %d", b.Remaining())
	}
}

func TestBudget_Apply_TruncatesEvents(t *testing.T) {
	b := NewBudget(2)
	events := []ChangeEvent{
		makeBudgetEvent(80),
		makeBudgetEvent(443),
		makeBudgetEvent(8080),
	}
	out := b.Apply(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events from Apply, got %d", len(out))
	}
	if out[0].Entry.LocalPort != 80 || out[1].Entry.LocalPort != 443 {
		t.Error("Apply returned unexpected events")
	}
}

func TestBudget_Apply_UnlimitedPassesAll(t *testing.T) {
	b := NewBudget(0)
	events := make([]ChangeEvent, 50)
	for i := range events {
		events[i] = makeBudgetEvent(uint16(1024 + i))
	}
	out := b.Apply(events)
	if len(out) != 50 {
		t.Fatalf("unlimited budget should pass all 50 events, got %d", len(out))
	}
}
