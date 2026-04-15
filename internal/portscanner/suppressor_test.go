package portscanner

import (
	"testing"
	"time"
)

func makeSuppEvent(proto, addr string, kind ChangeKind) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Protocol:    proto,
			LocalAddr:   addr,
			LocalPort:   8080,
			ProcessName: "svc",
		},
		Kind: kind,
	}
}

func TestSuppressor_FirstEventNotSuppressed(t *testing.T) {
	s := NewSuppressor(10 * time.Second)
	e := makeSuppEvent("tcp", "0.0.0.0", ChangeOpened)
	if s.IsSuppressed(dedupKey(e)) {
		t.Fatal("first event should not be suppressed")
	}
}

func TestSuppressor_SameKeyWithinWindow_IsSuppressed(t *testing.T) {
	s := NewSuppressor(10 * time.Second)
	e := makeSuppEvent("tcp", "0.0.0.0", ChangeOpened)
	key := dedupKey(e)
	s.IsSuppressed(key) // record
	if !s.IsSuppressed(key) {
		t.Fatal("second call within window should be suppressed")
	}
}

func TestSuppressor_SameKeyAfterWindow_NotSuppressed(t *testing.T) {
	fixed := time.Now()
	s := NewSuppressor(5 * time.Second)
	s.now = func() time.Time { return fixed }

	e := makeSuppEvent("tcp", "0.0.0.0", ChangeOpened)
	key := dedupKey(e)
	s.IsSuppressed(key) // record at fixed

	s.now = func() time.Time { return fixed.Add(6 * time.Second) }
	if s.IsSuppressed(key) {
		t.Fatal("event after quiet window should not be suppressed")
	}
}

func TestSuppressor_Flush_RemovesExpired(t *testing.T) {
	fixed := time.Now()
	s := NewSuppressor(2 * time.Second)
	s.now = func() time.Time { return fixed }

	e := makeSuppEvent("udp", "127.0.0.1", ChangeOpened)
	s.IsSuppressed(dedupKey(e))

	s.now = func() time.Time { return fixed.Add(3 * time.Second) }
	s.Flush()

	if len(s.suppressed) != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", len(s.suppressed))
	}
}

func TestSuppressor_Filter_RemovesSuppressed(t *testing.T) {
	s := NewSuppressor(30 * time.Second)
	e1 := makeSuppEvent("tcp", "0.0.0.0", ChangeOpened)
	e2 := makeSuppEvent("udp", "0.0.0.0", ChangeOpened)

	// pre-suppress e1
	s.IsSuppressed(dedupKey(e1))

	result := s.Filter([]ChangeEvent{e1, e2})
	if len(result) != 1 {
		t.Fatalf("expected 1 event after filter, got %d", len(result))
	}
	if result[0].Entry.Protocol != "udp" {
		t.Errorf("expected udp event to pass through, got %s", result[0].Entry.Protocol)
	}
}

func TestSuppressor_ZeroWindow_NeverSuppresses(t *testing.T) {
	s := NewSuppressor(0)
	e := makeSuppEvent("tcp", "0.0.0.0", ChangeOpened)
	key := dedupKey(e)
	s.IsSuppressed(key)
	if s.IsSuppressed(key) {
		t.Fatal("zero window should never suppress")
	}
}
