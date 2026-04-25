package portscanner

import (
	"testing"
	"time"
)

func makeReaperNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestReaper_TouchAndGet(t *testing.T) {
	base := time.Now()
	r := NewReaper(DefaultReaperPolicy())
	r.now = makeReaperNow(base)
	r.Touch("tcp:80")
	e, ok := r.Get("tcp:80")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.EventCount != 1 {
		t.Errorf("expected EventCount=1, got %d", e.EventCount)
	}
}

func TestReaper_TouchIncrements(t *testing.T) {
	base := time.Now()
	r := NewReaper(DefaultReaperPolicy())
	r.now = makeReaperNow(base)
	r.Touch("tcp:80")
	r.Touch("tcp:80")
	e, _ := r.Get("tcp:80")
	if e.EventCount != 2 {
		t.Errorf("expected EventCount=2, got %d", e.EventCount)
	}
}

func TestReaper_Reap_RemovesOldEntries(t *testing.T) {
	base := time.Now()
	current := base
	r := NewReaper(ReaperPolicy{MaxAge: 5 * time.Minute, Interval: time.Minute})
	r.now = func() time.Time { return current }
	r.Touch("tcp:80")
	current = base.Add(6 * time.Minute)
	reaped := r.Reap()
	if len(reaped) != 1 || reaped[0] != "tcp:80" {
		t.Errorf("expected tcp:80 to be reaped, got %v", reaped)
	}
	if r.Len() != 0 {
		t.Errorf("expected Len=0 after reap, got %d", r.Len())
	}
}

func TestReaper_Reap_KeepsFreshEntries(t *testing.T) {
	base := time.Now()
	current := base
	r := NewReaper(ReaperPolicy{MaxAge: 5 * time.Minute, Interval: time.Minute})
	r.now = func() time.Time { return current }
	r.Touch("tcp:443")
	current = base.Add(3 * time.Minute)
	reaped := r.Reap()
	if len(reaped) != 0 {
		t.Errorf("expected nothing reaped, got %v", reaped)
	}
	if r.Len() != 1 {
		t.Errorf("expected Len=1, got %d", r.Len())
	}
}

func TestReaper_MissingKey_ReturnsFalse(t *testing.T) {
	r := NewReaper(DefaultReaperPolicy())
	_, ok := r.Get("tcp:9999")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestReaper_DefaultPolicy(t *testing.T) {
	p := DefaultReaperPolicy()
	if p.MaxAge <= 0 {
		t.Error("expected positive MaxAge")
	}
	if p.Interval <= 0 {
		t.Error("expected positive Interval")
	}
}
