package portscanner

import (
	"testing"
	"time"
)

func makeTrackerNow(base time.Time) func() time.Time {
	current := base
	return func() time.Time { return current }
}

func TestStateChangeTracker_RecordFirstTime(t *testing.T) {
	base := time.Unix(1000, 0)
	nowFn := makeTrackerNow(base)
	tr := NewStateChangeTracker(nowFn)

	first, dur := tr.Record("tcp:80")
	if !first.Equal(base) {
		t.Errorf("expected first=%v got %v", base, first)
	}
	if dur != 0 {
		t.Errorf("expected zero duration, got %v", dur)
	}
}

func TestStateChangeTracker_DurationGrows(t *testing.T) {
	base := time.Unix(1000, 0)
	current := base
	nowFn := func() time.Time { return current }
	tr := NewStateChangeTracker(nowFn)

	tr.Record("tcp:443")
	current = base.Add(5 * time.Second)
	_, dur := tr.Record("tcp:443")
	if dur != 5*time.Second {
		t.Errorf("expected 5s duration, got %v", dur)
	}
}

func TestStateChangeTracker_FirstSeenUnchanged(t *testing.T) {
	base := time.Unix(2000, 0)
	current := base
	nowFn := func() time.Time { return current }
	tr := NewStateChangeTracker(nowFn)

	tr.Record("udp:53")
	current = base.Add(10 * time.Second)
	tr.Record("udp:53")

	first, ok := tr.FirstSeen("udp:53")
	if !ok || !first.Equal(base) {
		t.Errorf("expected first=%v ok=true, got first=%v ok=%v", base, first, ok)
	}
}

func TestStateChangeTracker_Forget(t *testing.T) {
	base := time.Unix(3000, 0)
	tr := NewStateChangeTracker(func() time.Time { return base })
	tr.Record("tcp:22")
	tr.Forget("tcp:22")

	_, ok := tr.FirstSeen("tcp:22")
	if ok {
		t.Error("expected key to be forgotten")
	}
	if tr.Len() != 0 {
		t.Errorf("expected len=0, got %d", tr.Len())
	}
}

func TestStateChangeTracker_IndependentKeys(t *testing.T) {
	base := time.Unix(4000, 0)
	current := base
	nowFn := func() time.Time { return current }
	tr := NewStateChangeTracker(nowFn)

	tr.Record("tcp:80")
	current = base.Add(3 * time.Second)
	tr.Record("tcp:8080")

	f1, _ := tr.FirstSeen("tcp:80")
	f2, _ := tr.FirstSeen("tcp:8080")
	if f1.Equal(f2) {
		t.Error("expected independent first-seen times")
	}
	if tr.Len() != 2 {
		t.Errorf("expected len=2, got %d", tr.Len())
	}
}
