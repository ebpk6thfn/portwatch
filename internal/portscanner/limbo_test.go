package portscanner

import (
	"testing"
	"time"
)

func makeLimboEvent(port uint16, proto, process string, kind ChangeKind) ChangeEvent {
	return ChangeEvent{
		Kind: kind,
		Entry: Entry{
			Port:     port,
			Protocol: proto,
			Process:  process,
		},
	}
}

func TestLimbo_HoldNewEvent_ReturnsTrue(t *testing.T) {
	now := time.Now()
	l := NewLimbo(DefaultLimboPolicy(), func() time.Time { return now })

	ev := makeLimboEvent(8080, "tcp", "nginx", ChangeOpened)
	if !l.Hold(ev) {
		t.Fatal("expected Hold to return true for new event")
	}
	if l.Len() != 1 {
		t.Fatalf("expected len 1, got %d", l.Len())
	}
}

func TestLimbo_HoldSameEvent_ReturnsFalse(t *testing.T) {
	now := time.Now()
	l := NewLimbo(DefaultLimboPolicy(), func() time.Time { return now })

	ev := makeLimboEvent(8080, "tcp", "nginx", ChangeOpened)
	l.Hold(ev)
	if l.Hold(ev) {
		t.Fatal("expected Hold to return false for confirmed event")
	}
	if l.Len() != 0 {
		t.Fatalf("expected len 0 after confirmation, got %d", l.Len())
	}
}

func TestLimbo_Flush_ExpiredEvents(t *testing.T) {
	base := time.Now()
	current := base
	l := NewLimbo(LimboPolicy{Window: 2 * time.Second, MaxSize: 64}, func() time.Time { return current })

	ev := makeLimboEvent(9090, "tcp", "sshd", ChangeOpened)
	l.Hold(ev)

	// advance time past window
	current = base.Add(3 * time.Second)
	expired := l.Flush()
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired event, got %d", len(expired))
	}
	if l.Len() != 0 {
		t.Fatal("expected empty limbo after flush")
	}
}

func TestLimbo_Flush_NotExpiredYet(t *testing.T) {
	base := time.Now()
	current := base
	l := NewLimbo(LimboPolicy{Window: 10 * time.Second, MaxSize: 64}, func() time.Time { return current })

	ev := makeLimboEvent(3000, "tcp", "app", ChangeOpened)
	l.Hold(ev)

	current = base.Add(1 * time.Second)
	expired := l.Flush()
	if len(expired) != 0 {
		t.Fatalf("expected 0 expired events, got %d", len(expired))
	}
	if l.Len() != 1 {
		t.Fatal("expected event still in limbo")
	}
}

func TestLimbo_MaxSize_EvictsOldest(t *testing.T) {
	base := time.Now()
	current := base
	l := NewLimbo(LimboPolicy{Window: 60 * time.Second, MaxSize: 2}, func() time.Time { return current })

	l.Hold(makeLimboEvent(1001, "tcp", "a", ChangeOpened))
	current = base.Add(1 * time.Second)
	l.Hold(makeLimboEvent(1002, "tcp", "b", ChangeOpened))
	current = base.Add(2 * time.Second)
	l.Hold(makeLimboEvent(1003, "tcp", "c", ChangeOpened))

	if l.Len() != 2 {
		t.Fatalf("expected len 2 after eviction, got %d", l.Len())
	}
}

func TestLimbo_DifferentPortsAreIndependent(t *testing.T) {
	now := time.Now()
	l := NewLimbo(DefaultLimboPolicy(), func() time.Time { return now })

	ev1 := makeLimboEvent(80, "tcp", "nginx", ChangeOpened)
	ev2 := makeLimboEvent(443, "tcp", "nginx", ChangeOpened)

	l.Hold(ev1)
	l.Hold(ev2)

	if l.Len() != 2 {
		t.Fatalf("expected 2 independent entries, got %d", l.Len())
	}
}
