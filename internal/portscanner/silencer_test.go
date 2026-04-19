package portscanner

import (
	"testing"
	"time"
)

func makeSilencerNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeSilenceEvent(port uint16) ChangeEvent {
	return ChangeEvent{Entry: Entry{Port: port, Protocol: "tcp"}}
}

func TestSilencer_NotSilencedByDefault(t *testing.T) {
	now := time.Now()
	s := NewSilencer(makeSilencerNow(now))
	if s.IsSilenced(8080) {
		t.Fatal("expected port not silenced")
	}
}

func TestSilencer_SilencedWithinWindow(t *testing.T) {
	now := time.Now()
	s := NewSilencer(makeSilencerNow(now))
	s.Silence(8080, now.Add(10*time.Minute))
	if !s.IsSilenced(8080) {
		t.Fatal("expected port to be silenced")
	}
}

func TestSilencer_NotSilencedAfterExpiry(t *testing.T) {
	now := time.Now()
	s := NewSilencer(makeSilencerNow(now))
	s.Silence(8080, now.Add(-1*time.Second))
	if s.IsSilenced(8080) {
		t.Fatal("expected silence to have expired")
	}
}

func TestSilencer_Filter_DropsMatchingPort(t *testing.T) {
	now := time.Now()
	s := NewSilencer(makeSilencerNow(now))
	s.Silence(9000, now.Add(5*time.Minute))
	events := []ChangeEvent{
		makeSilenceEvent(9000),
		makeSilenceEvent(8080),
	}
	out := s.Filter(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
	if out[0].Entry.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", out[0].Entry.Port)
	}
}

func TestSilencer_Flush_RemovesExpired(t *testing.T) {
	now := time.Now()
	s := NewSilencer(makeSilencerNow(now))
	s.Silence(1234, now.Add(-1*time.Second))
	s.Silence(5678, now.Add(10*time.Minute))
	s.Flush()
	if s.IsSilenced(1234) {
		t.Fatal("expected expired rule to be flushed")
	}
	if !s.IsSilenced(5678) {
		t.Fatal("expected active rule to remain")
	}
}

func TestSilencer_IndependentPorts(t *testing.T) {
	now := time.Now()
	s := NewSilencer(makeSilencerNow(now))
	s.Silence(443, now.Add(1*time.Hour))
	if s.IsSilenced(80) {
		t.Fatal("port 80 should not be silenced")
	}
}
