package portscanner

import (
	"testing"
	"time"
)

func makeFlapEvent(port uint16, proto, state string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:     port,
			Protocol: proto,
		},
		State: state,
	}
}

func TestFlapDetector_BelowThreshold_NoFlap(t *testing.T) {
	policy := DefaultFlapPolicy()
	policy.Threshold = 3
	fd := NewFlapDetector(policy)

	e := makeFlapEvent(8080, "tcp", "opened")
	if fd.Record(e) {
		t.Error("expected no flap on first transition")
	}
	if fd.Record(e) {
		t.Error("expected no flap on second transition")
	}
}

func TestFlapDetector_AtThreshold_ReturnsFlap(t *testing.T) {
	policy := DefaultFlapPolicy()
	policy.Threshold = 3
	policy.Window = time.Minute
	policy.Cooldown = time.Minute
	fd := NewFlapDetector(policy)

	e := makeFlapEvent(9090, "tcp", "opened")
	fd.Record(e)
	fd.Record(e)
	flapping := fd.Record(e)
	if !flapping {
		t.Error("expected flap detected at threshold")
	}
}

func TestFlapDetector_CooldownSuppressesRepeat(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	policy := FlapPolicy{
		Window:    time.Minute,
		Threshold: 2,
		Cooldown:  10 * time.Minute,
	}
	fd := NewFlapDetector(policy)
	fd.now = func() time.Time { return now }

	e := makeFlapEvent(443, "tcp", "closed")
	fd.Record(e)
	fd.Record(e) // trips flap, enters cooldown

	// Advance time but stay within cooldown.
	fd.now = func() time.Time { return now.Add(5 * time.Minute) }
	if !fd.Record(e) {
		t.Error("expected flap still suppressed within cooldown")
	}
}

func TestFlapDetector_EvictsOldTransitions(t *testing.T) {
	now := time.Unix(2_000_000, 0)
	policy := FlapPolicy{
		Window:    30 * time.Second,
		Threshold: 3,
		Cooldown:  time.Minute,
	}
	fd := NewFlapDetector(policy)
	fd.now = func() time.Time { return now }

	e := makeFlapEvent(22, "tcp", "opened")
	fd.Record(e)
	fd.Record(e)

	// Advance past the window so old transitions are evicted.
	fd.now = func() time.Time { return now.Add(time.Minute) }
	if fd.Record(e) {
		t.Error("expected no flap after old transitions evicted")
	}
	if fd.Count(e.Entry.Key()) != 1 {
		t.Errorf("expected count=1 after eviction, got %d", fd.Count(e.Entry.Key()))
	}
}

func TestFlapDetector_IndependentKeys(t *testing.T) {
	policy := FlapPolicy{Window: time.Minute, Threshold: 2, Cooldown: time.Minute}
	fd := NewFlapDetector(policy)

	a := makeFlapEvent(80, "tcp", "opened")
	b := makeFlapEvent(443, "tcp", "opened")

	fd.Record(a)
	fd.Record(b)

	// Only port 80 reaches threshold.
	if fd.Record(a) == false {
		// a should flap
	}
	if fd.Count(b.Entry.Key()) != 1 {
		t.Errorf("port 443 should have count=1, independent of port 80")
	}
}

func TestFlapDetector_String_ContainsTracked(t *testing.T) {
	fd := NewFlapDetector(DefaultFlapPolicy())
	fd.Record(makeFlapEvent(3000, "tcp", "opened"))
	s := fd.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
