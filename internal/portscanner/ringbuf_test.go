package portscanner

import (
	"testing"
)

func makeRBEvent(port uint16) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{Port: port, Protocol: "tcp"},
		Type:  EventOpened,
	}
}

func TestRingBuffer_EmptyDrain(t *testing.T) {
	rb := NewRingBuffer(4)
	if got := rb.Drain(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestRingBuffer_PushAndDrain(t *testing.T) {
	rb := NewRingBuffer(4)
	rb.Push(makeRBEvent(80))
	rb.Push(makeRBEvent(443))
	events := rb.Drain()
	if len(events) != 2 {
		t.Fatalf("expected 2, got %d", len(events))
	}
	if events[0].Entry.Port != 80 || events[1].Entry.Port != 443 {
		t.Errorf("unexpected order: %v", events)
	}
}

func TestRingBuffer_DrainClearsBuffer(t *testing.T) {
	rb := NewRingBuffer(4)
	rb.Push(makeRBEvent(8080))
	rb.Drain()
	if rb.Len() != 0 {
		t.Fatalf("expected 0 after drain, got %d", rb.Len())
	}
}

func TestRingBuffer_Overflow_OldestOverwritten(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Push(makeRBEvent(1))
	rb.Push(makeRBEvent(2))
	rb.Push(makeRBEvent(3))
	rb.Push(makeRBEvent(4)) // overwrites port 1
	events := rb.Drain()
	if len(events) != 3 {
		t.Fatalf("expected 3, got %d", len(events))
	}
	if events[0].Entry.Port != 2 {
		t.Errorf("expected oldest=2, got %d", events[0].Entry.Port)
	}
	if events[2].Entry.Port != 4 {
		t.Errorf("expected newest=4, got %d", events[2].Entry.Port)
	}
}

func TestRingBuffer_LenAndCap(t *testing.T) {
	rb := NewRingBuffer(5)
	if rb.Cap() != 5 {
		t.Fatalf("expected cap 5")
	}
	rb.Push(makeRBEvent(9000))
	if rb.Len() != 1 {
		t.Fatalf("expected len 1")
	}
}

func TestRingBuffer_ZeroCapacity_ClampedToOne(t *testing.T) {
	rb := NewRingBuffer(0)
	if rb.Cap() != 1 {
		t.Fatalf("expected cap clamped to 1")
	}
	rb.Push(makeRBEvent(22))
	rb.Push(makeRBEvent(23))
	events := rb.Drain()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}
