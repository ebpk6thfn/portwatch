package portscanner

import (
	"testing"
	"time"
)

func makePausableEvent(port uint16) ChangeEvent {
	return ChangeEvent{
		Entry:     Entry{Port: port, Protocol: "tcp"},
		Type:      EventOpened,
		Timestamp: time.Now(),
	}
}

func TestPausable_InitiallyRunning(t *testing.T) {
	p := NewPausable()
	if p.IsPaused() {
		t.Fatal("expected not paused initially")
	}
}

func TestPausable_PauseStopsEvents(t *testing.T) {
	p := NewPausable()
	p.Pause()
	events := []ChangeEvent{makePausableEvent(8080)}
	out := p.Filter(events)
	if len(out) != 0 {
		t.Fatalf("expected 0 events while paused, got %d", len(out))
	}
}

func TestPausable_ResumeRestoresEvents(t *testing.T) {
	p := NewPausable()
	p.Pause()
	p.Resume()
	events := []ChangeEvent{makePausableEvent(8080), makePausableEvent(9090)}
	out := p.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events after resume, got %d", len(out))
	}
}

func TestPausable_IsPaused_Toggles(t *testing.T) {
	p := NewPausable()
	p.Pause()
	if !p.IsPaused() {
		t.Fatal("expected paused")
	}
	p.Resume()
	if p.IsPaused() {
		t.Fatal("expected not paused after resume")
	}
}

func TestPausable_Filter_EmptyInput(t *testing.T) {
	p := NewPausable()
	out := p.Filter(nil)
	if out != nil {
		t.Fatalf("expected nil for nil input, got %v", out)
	}
}
