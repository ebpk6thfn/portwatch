package portscanner

import (
	"net"
	"testing"
	"time"
)

func makeDedupEvent(proto, ip string, port uint16, evType string) ChangeEvent {
	return ChangeEvent{
		Type: evType,
		Entry: Entry{
			Protocol:  proto,
			LocalIP:   net.ParseIP(ip),
			LocalPort: port,
		},
	}
}

func TestDeduplicator_ZeroWindow_NeverDeduplicates(t *testing.T) {
	d := NewDeduplicator(0)
	e := makeDedupEvent("tcp", "0.0.0.0", 8080, "opened")

	if d.IsDuplicate(e) {
		t.Fatal("expected false for zero-window deduplicator")
	}
	if d.IsDuplicate(e) {
		t.Fatal("expected false on second call with zero window")
	}
}

func TestDeduplicator_FirstEventNotDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	e := makeDedupEvent("tcp", "0.0.0.0", 9090, "opened")

	if d.IsDuplicate(e) {
		t.Fatal("first event should not be a duplicate")
	}
}

func TestDeduplicator_SameEventWithinWindow_IsDuplicate(t *testing.T) {
	window := 30 * time.Second
	now := time.Now()
	d := NewDeduplicator(window)
	d.now = func() time.Time { return now }

	e := makeDedupEvent("tcp", "0.0.0.0", 8080, "opened")
	d.IsDuplicate(e) // record it

	if !d.IsDuplicate(e) {
		t.Fatal("same event within window should be duplicate")
	}
}

func TestDeduplicator_SameEventAfterWindow_NotDuplicate(t *testing.T) {
	window := 5 * time.Second
	now := time.Now()
	d := NewDeduplicator(window)
	d.now = func() time.Time { return now }

	e := makeDedupEvent("tcp", "127.0.0.1", 3000, "closed")
	d.IsDuplicate(e) // record

	d.now = func() time.Time { return now.Add(6 * time.Second) }
	if d.IsDuplicate(e) {
		t.Fatal("event after window expiry should not be duplicate")
	}
}

func TestDeduplicator_DifferentType_NotDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	opened := makeDedupEvent("tcp", "0.0.0.0", 8080, "opened")
	closed := makeDedupEvent("tcp", "0.0.0.0", 8080, "closed")

	d.IsDuplicate(opened)
	if d.IsDuplicate(closed) {
		t.Fatal("different event types should not be considered duplicates")
	}
}

func TestDeduplicator_Filter_RemovesDuplicates(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(30 * time.Second)
	d.now = func() time.Time { return now }

	e1 := makeDedupEvent("tcp", "0.0.0.0", 8080, "opened")
	e2 := makeDedupEvent("udp", "0.0.0.0", 5353, "opened")

	events := []ChangeEvent{e1, e2, e1} // e1 duplicated
	result := d.Filter(events)

	if len(result) != 2 {
		t.Fatalf("expected 2 events after dedup, got %d", len(result))
	}
}

func TestDeduplicator_Purge_RemovesStaleEntries(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(5 * time.Second)
	d.now = func() time.Time { return now }

	e := makeDedupEvent("tcp", "0.0.0.0", 443, "opened")
	d.IsDuplicate(e)

	d.now = func() time.Time { return now.Add(10 * time.Second) }
	d.Purge()

	if len(d.seen) != 0 {
		t.Fatalf("expected seen map to be empty after purge, got %d entries", len(d.seen))
	}
}
