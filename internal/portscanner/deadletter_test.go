package portscanner

import (
	"testing"
	"time"
)

func makeDLEvent(port uint16) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			LocalPort: port,
			Protocol:  "tcp",
		},
		Type:      EventOpened,
		Timestamp: time.Now(),
	}
}

func TestDeadLetterQueue_PushAndLen(t *testing.T) {
	q := NewDeadLetterQueue(10)
	q.Push(makeDLEvent(8080), ReasonQuotaExceeded)
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestDeadLetterQueue_DrainClearsBuffer(t *testing.T) {
	q := NewDeadLetterQueue(10)
	q.Push(makeDLEvent(8080), ReasonSuppressed)
	q.Push(makeDLEvent(9090), ReasonCircuitOpen)
	items := q.Drain()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if q.Len() != 0 {
		t.Fatalf("expected queue empty after drain, got %d", q.Len())
	}
}

func TestDeadLetterQueue_Overflow_EvictsOldest(t *testing.T) {
	q := NewDeadLetterQueue(3)
	q.Push(makeDLEvent(1), ReasonQuotaExceeded)
	q.Push(makeDLEvent(2), ReasonSuppressed)
	q.Push(makeDLEvent(3), ReasonCircuitOpen)
	q.Push(makeDLEvent(4), ReasonDeliveryFailed)

	if q.Len() != 3 {
		t.Fatalf("expected len 3, got %d", q.Len())
	}
	items := q.Drain()
	if items[0].Event.Entry.LocalPort != 2 {
		t.Errorf("expected oldest evicted; first item port = %d", items[0].Event.Entry.LocalPort)
	}
}

func TestDeadLetterQueue_CountByReason(t *testing.T) {
	q := NewDeadLetterQueue(20)
	q.Push(makeDLEvent(80), ReasonQuotaExceeded)
	q.Push(makeDLEvent(81), ReasonQuotaExceeded)
	q.Push(makeDLEvent(82), ReasonSuppressed)

	counts := q.CountByReason()
	if counts[ReasonQuotaExceeded] != 2 {
		t.Errorf("expected 2 quota_exceeded, got %d", counts[ReasonQuotaExceeded])
	}
	if counts[ReasonSuppressed] != 1 {
		t.Errorf("expected 1 suppressed, got %d", counts[ReasonSuppressed])
	}
	if counts[ReasonCircuitOpen] != 0 {
		t.Errorf("expected 0 circuit_open, got %d", counts[ReasonCircuitOpen])
	}
}

func TestDeadLetterQueue_DrainPreservesReason(t *testing.T) {
	q := NewDeadLetterQueue(10)
	q.Push(makeDLEvent(443), ReasonDeliveryFailed)
	items := q.Drain()
	if items[0].Reason != ReasonDeliveryFailed {
		t.Errorf("expected reason %q, got %q", ReasonDeliveryFailed, items[0].Reason)
	}
}

func TestDeadLetterQueue_ZeroMaxSize_UsesDefault(t *testing.T) {
	q := NewDeadLetterQueue(0)
	if q.maxSize != 256 {
		t.Errorf("expected default maxSize 256, got %d", q.maxSize)
	}
}
