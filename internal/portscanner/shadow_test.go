package portscanner

import (
	"testing"
	"time"
)

func makeShadowEvent(port uint16, kind string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:     port,
			Protocol: "tcp",
			IP:       "127.0.0.1",
		},
		Kind:      kind,
		Timestamp: time.Now(),
	}
}

func TestShadow_Disabled_PassesEvent(t *testing.T) {
	p := DefaultShadowPolicy()
	p.Enabled = false
	s := NewShadow(p)

	ev := makeShadowEvent(8080, "opened")
	out := s.Filter(ev)
	if out == nil {
		t.Fatal("expected event to pass through when shadow disabled")
	}
	if out.Entry.Port != 8080 {
		t.Errorf("unexpected port: got %d", out.Entry.Port)
	}
}

func TestShadow_Enabled_SuppressesEvent(t *testing.T) {
	p := DefaultShadowPolicy()
	p.Enabled = true
	s := NewShadow(p)

	ev := makeShadowEvent(443, "opened")
	out := s.Filter(ev)
	if out != nil {
		t.Fatal("expected nil when shadow mode enabled")
	}
}

func TestShadow_LogDropped_RecordsEvent(t *testing.T) {
	p := ShadowPolicy{Enabled: true, LogDropped: true, MaxDropped: 10}
	s := NewShadow(p)

	s.Filter(makeShadowEvent(80, "opened"))
	s.Filter(makeShadowEvent(443, "opened"))

	if s.Len() != 2 {
		t.Fatalf("expected 2 dropped events, got %d", s.Len())
	}
}

func TestShadow_LogDropped_False_DoesNotRecord(t *testing.T) {
	p := ShadowPolicy{Enabled: true, LogDropped: false, MaxDropped: 10}
	s := NewShadow(p)

	s.Filter(makeShadowEvent(80, "opened"))
	if s.Len() != 0 {
		t.Fatalf("expected 0 dropped events, got %d", s.Len())
	}
}

func TestShadow_MaxDropped_EvictsOldest(t *testing.T) {
	p := ShadowPolicy{Enabled: true, LogDropped: true, MaxDropped: 3}
	s := NewShadow(p)

	for i := uint16(1); i <= 5; i++ {
		s.Filter(makeShadowEvent(i, "opened"))
	}

	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}
	dropped := s.Dropped()
	if dropped[0].Event.Entry.Port != 3 {
		t.Errorf("expected oldest retained port=3, got %d", dropped[0].Event.Entry.Port)
	}
}

func TestShadow_Clear_RemovesAll(t *testing.T) {
	p := ShadowPolicy{Enabled: true, LogDropped: true, MaxDropped: 10}
	s := NewShadow(p)

	s.Filter(makeShadowEvent(22, "opened"))
	s.Clear()

	if s.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", s.Len())
	}
}

func TestShadow_IsEnabled(t *testing.T) {
	s := NewShadow(ShadowPolicy{Enabled: true})
	if !s.IsEnabled() {
		t.Error("expected IsEnabled to return true")
	}

	s2 := NewShadow(ShadowPolicy{Enabled: false})
	if s2.IsEnabled() {
		t.Error("expected IsEnabled to return false")
	}
}
