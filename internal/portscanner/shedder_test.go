package portscanner

import (
	"testing"
	"time"
)

func makeShedderEvent(port uint16) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{Port: port, Protocol: "tcp"},
		Type:  EventOpened,
	}
}

func TestShedder_BelowDepth_AllowsAll(t *testing.T) {
	p := DefaultShedderPolicy()
	s := NewShedder(p)
	s.SetDepth(10) // well below 200

	for i := 0; i < 20; i++ {
		if !s.Allow(makeShedderEvent(uint16(8000 + i))) {
			t.Errorf("expected Allow=true at low depth, i=%d", i)
		}
	}
}

func TestShedder_AtDepth_EntersShedding(t *testing.T) {
	p := DefaultShedderPolicy()
	p.MaxQueueDepth = 5
	p.ShedPercent = 0.5
	p.CooldownPeriod = time.Minute
	s := NewShedder(p)
	s.SetDepth(5)

	if !s.IsShedding() {
		// IsShedding is only set after the first Allow call when depth>=max
	}

	// After an Allow call the shedder should activate.
	s.Allow(makeShedderEvent(9000))
	if !s.IsShedding() {
		t.Fatal("expected shedder to be active after depth >= MaxQueueDepth")
	}
}

func TestShedder_ShedPercent100_DropsAll(t *testing.T) {
	p := ShedderPolicy{
		MaxQueueDepth:  1,
		ShedPercent:    1.0,
		CooldownPeriod: time.Minute,
	}
	s := NewShedder(p)
	s.SetDepth(1)

	for i := 0; i < 10; i++ {
		if s.Allow(makeShedderEvent(uint16(7000 + i))) {
			t.Errorf("expected Allow=false with 100%% shedding, i=%d", i)
		}
	}
}

func TestShedder_CooldownLifts(t *testing.T) {
	now := time.Now()
	p := ShedderPolicy{
		MaxQueueDepth:  1,
		ShedPercent:    1.0,
		CooldownPeriod: 5 * time.Second,
	}
	s := NewShedder(p)
	s.now = func() time.Time { return now }
	s.SetDepth(1)
	s.Allow(makeShedderEvent(9000)) // triggers shedding

	if !s.IsShedding() {
		t.Fatal("expected shedding to be active")
	}

	// Advance time past cooldown with depth back to zero.
	s.SetDepth(0)
	s.now = func() time.Time { return now.Add(10 * time.Second) }
	s.Allow(makeShedderEvent(9001)) // should clear shedding

	if s.IsShedding() {
		t.Fatal("expected shedding to be cleared after cooldown")
	}
}

func TestShedder_Filter_ReducesSlice(t *testing.T) {
	p := ShedderPolicy{
		MaxQueueDepth:  1,
		ShedPercent:    0.5,
		CooldownPeriod: time.Minute,
	}
	s := NewShedder(p)
	s.SetDepth(1)

	events := make([]ChangeEvent, 10)
	for i := range events {
		events[i] = makeShedderEvent(uint16(6000 + i))
	}

	out := s.Filter(events)
	if len(out) == 0 {
		t.Fatal("expected some events to pass through")
	}
	if len(out) >= len(events) {
		t.Errorf("expected fewer events after shedding, got %d/%d", len(out), len(events))
	}
}
