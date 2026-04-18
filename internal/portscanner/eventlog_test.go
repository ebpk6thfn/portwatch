package portscanner

import (
	"testing"
	"time"
)

func makeLogEvent(port uint16, kind string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{Port: port, Protocol: "tcp"},
		Kind:  kind,
	}
}

func TestEventLog_EmptyLen(t *testing.T) {
	l := NewEventLog(10)
	if l.Len() != 0 {
		t.Fatalf("expected 0, got %d", l.Len())
	}
}

func TestEventLog_RecordAndAll(t *testing.T) {
	l := NewEventLog(10)
	l.Record(makeLogEvent(80, "opened"))
	l.Record(makeLogEvent(443, "opened"))
	if l.Len() != 2 {
		t.Fatalf("expected 2, got %d", l.Len())
	}
	all := l.All()
	if all[0].Event.Entry.Port != 80 {
		t.Errorf("expected port 80, got %d", all[0].Event.Entry.Port)
	}
	if all[1].Event.Entry.Port != 443 {
		t.Errorf("expected port 443, got %d", all[1].Event.Entry.Port)
	}
}

func TestEventLog_Overflow_EvictsOldest(t *testing.T) {
	l := NewEventLog(3)
	l.Record(makeLogEvent(1, "opened"))
	l.Record(makeLogEvent(2, "opened"))
	l.Record(makeLogEvent(3, "opened"))
	l.Record(makeLogEvent(4, "opened"))
	if l.Len() != 3 {
		t.Fatalf("expected 3, got %d", l.Len())
	}
	all := l.All()
	if all[0].Event.Entry.Port != 2 {
		t.Errorf("expected oldest evicted; first port should be 2, got %d", all[0].Event.Entry.Port)
	}
}

func TestEventLog_Clear(t *testing.T) {
	l := NewEventLog(10)
	l.Record(makeLogEvent(80, "opened"))
	l.Clear()
	if l.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", l.Len())
	}
}

func TestEventLog_Since(t *testing.T) {
	l := NewEventLog(10)
	before := time.Now()
	time.Sleep(2 * time.Millisecond)
	l.Record(makeLogEvent(80, "opened"))
	l.Record(makeLogEvent(443, "opened"))

	results := l.Since(before)
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}

	future := time.Now().Add(time.Hour)
	results = l.Since(future)
	if len(results) != 0 {
		t.Fatalf("expected 0 for future cutoff, got %d", len(results))
	}
}
