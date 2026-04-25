package portscanner

import (
	"fmt"
	"testing"
	"time"
)

func makeLedgerEvent(port int, proto, process string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:     port,
			Protocol: proto,
			Process:  process,
			Addr:     "0.0.0.0",
		},
		Type: EventOpened,
	}
}

func TestLedger_RecordAndGet(t *testing.T) {
	l := NewLedger(0)
	now := time.Now()
	e := makeLedgerEvent(8080, "tcp", "nginx")
	l.Record(e, now)

	entry, ok := l.Get(e.Entry.Key())
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if entry.Port != 8080 {
		t.Errorf("expected port 8080, got %d", entry.Port)
	}
	if entry.Count != 1 {
		t.Errorf("expected count 1, got %d", entry.Count)
	}
	if !entry.FirstSeen.Equal(now) {
		t.Errorf("expected FirstSeen %v, got %v", now, entry.FirstSeen)
	}
}

func TestLedger_RecordUpdatesLastSeen(t *testing.T) {
	l := NewLedger(0)
	t0 := time.Now()
	t1 := t0.Add(5 * time.Second)
	e := makeLedgerEvent(9090, "tcp", "app")
	l.Record(e, t0)
	l.Record(e, t1)

	entry, ok := l.Get(e.Entry.Key())
	if !ok {
		t.Fatal("expected entry")
	}
	if entry.Count != 2 {
		t.Errorf("expected count 2, got %d", entry.Count)
	}
	if !entry.FirstSeen.Equal(t0) {
		t.Errorf("FirstSeen should remain %v, got %v", t0, entry.FirstSeen)
	}
	if !entry.LastSeen.Equal(t1) {
		t.Errorf("LastSeen should be %v, got %v", t1, entry.LastSeen)
	}
}

func TestLedger_MissingKey(t *testing.T) {
	l := NewLedger(0)
	_, ok := l.Get("nonexistent")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestLedger_All_ReturnsSnapshot(t *testing.T) {
	l := NewLedger(0)
	now := time.Now()
	for i := 0; i < 3; i++ {
		l.Record(makeLedgerEvent(8000+i, "tcp", fmt.Sprintf("proc%d", i)), now)
	}
	all := l.All()
	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}
}

func TestLedger_MaxSize_EvictsOldest(t *testing.T) {
	l := NewLedger(2)
	t0 := time.Unix(1000, 0)
	t1 := time.Unix(2000, 0)
	t2 := time.Unix(3000, 0)

	e0 := makeLedgerEvent(8001, "tcp", "a")
	e1 := makeLedgerEvent(8002, "tcp", "b")
	e2 := makeLedgerEvent(8003, "tcp", "c")

	l.Record(e0, t0)
	l.Record(e1, t1)
	l.Record(e2, t2)

	if l.Len() != 2 {
		t.Errorf("expected 2 entries after eviction, got %d", l.Len())
	}
	_, stillHasOldest := l.Get(e0.Entry.Key())
	if stillHasOldest {
		t.Error("oldest entry should have been evicted")
	}
}

func TestLedger_Len(t *testing.T) {
	l := NewLedger(0)
	if l.Len() != 0 {
		t.Error("expected initial len 0")
	}
	now := time.Now()
	l.Record(makeLedgerEvent(1234, "udp", "svc"), now)
	if l.Len() != 1 {
		t.Errorf("expected len 1, got %d", l.Len())
	}
}
