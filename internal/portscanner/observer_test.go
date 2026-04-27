package portscanner

import (
	"testing"
	"time"
)

func makeObserverEntry(port int, proto string) Entry {
	return Entry{
		LocalPort: port,
		Protocol:  proto,
		State:     "LISTEN",
	}
}

func fixedObserverNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestObserver_BelowThreshold_NoEvent(t *testing.T) {
	policy := ObserverPolicy{MinObservations: 3, Window: time.Minute}
	obs := NewObserver(policy)
	e := makeObserverEntry(8080, "tcp")

	for i := 0; i < 2; i++ {
		ev, ok := obs.Record(e)
		if ok || ev != nil {
			t.Fatalf("expected no event before threshold, got one on iteration %d", i)
		}
	}
}

func TestObserver_AtThreshold_ReturnsEvent(t *testing.T) {
	policy := ObserverPolicy{MinObservations: 3, Window: time.Minute}
	obs := NewObserver(policy)
	e := makeObserverEntry(9090, "tcp")

	var last *ObserverEvent
	for i := 0; i < 3; i++ {
		ev, ok := obs.Record(e)
		if i == 2 {
			if !ok || ev == nil {
				t.Fatal("expected event at threshold")
			}
			last = ev
		}
	}

	if last.SeenCount != 3 {
		t.Errorf("expected SeenCount=3, got %d", last.SeenCount)
	}
	if last.Entry.LocalPort != 9090 {
		t.Errorf("unexpected port in event: %d", last.Entry.LocalPort)
	}
}

func TestObserver_WindowExpiry_ResetsCount(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	policy := ObserverPolicy{MinObservations: 2, Window: time.Minute}
	obs := NewObserver(policy)
	obs.now = fixedObserverNow(base)

	e := makeObserverEntry(443, "tcp")
	obs.Record(e) // count = 1, no event

	// Advance past the window
	obs.now = fixedObserverNow(base.Add(2 * time.Minute))
	ev, ok := obs.Record(e) // should reset, count = 1 again
	if ok || ev != nil {
		t.Fatal("expected no event after window reset")
	}
}

func TestObserver_IndependentEntries(t *testing.T) {
	policy := ObserverPolicy{MinObservations: 2, Window: time.Minute}
	obs := NewObserver(policy)

	a := makeObserverEntry(80, "tcp")
	b := makeObserverEntry(443, "tcp")

	obs.Record(a)
	obs.Record(b)

	_, okA := obs.Record(a)
	_, okB := obs.Record(b)

	if !okA {
		t.Error("expected event for entry a")
	}
	if !okB {
		t.Error("expected event for entry b")
	}
}

func TestObserver_Flush_RemovesExpired(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	policy := ObserverPolicy{MinObservations: 5, Window: 30 * time.Second}
	obs := NewObserver(policy)
	obs.now = fixedObserverNow(base)

	e := makeObserverEntry(22, "tcp")
	obs.Record(e)

	if obs.Len() != 1 {
		t.Fatalf("expected 1 record before flush, got %d", obs.Len())
	}

	obs.now = fixedObserverNow(base.Add(time.Minute))
	obs.Flush()

	if obs.Len() != 0 {
		t.Errorf("expected 0 records after flush, got %d", obs.Len())
	}
}
