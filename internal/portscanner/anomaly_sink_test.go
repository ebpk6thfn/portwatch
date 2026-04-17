package portscanner

import (
	"testing"
	"time"
)

func TestAnomalySink_PushAndDrain(t *testing.T := NewAnomalySink(10)
	s.Push(Anomaly{Type: AnomalyBurst, Port: 80})
	s.Push(Anomaly{Type: AnomalyRapidCycle, Port: 443})

	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
	out := s.Drain()
	if len(out) != 2 {
		t.Fatalf("expected 2 drained, got %d", len(out))
	}
	if s.Len() != 0 {
		t.Fatal("sink should be empty after drain")
	}
}

func TestAnomalySink_Overflow_DropsOldest(t *testing.T) {
	s := NewAnomalySink(3)
	for i := uint16(1); i <= 4; i++ {
		s.Push(Anomaly{Type: AnomalyBurst, Port: i})
	}
	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}
	out := s.Drain()
	if out[0].Port != 2 {
		t.Fatalf("expected oldest dropped, first port should be 2, got %d", out[0].Port)
	}
}

func TestAnomalySink_DrainClearsBuffer(t *testing.T) {
	s := NewAnomalySink(10)
	s.Push(Anomaly{Type: AnomalyUnknown, Port: 9999})
	s.Drain()
	if s.Len() != 0 {
		t.Fatal("expected empty after drain")
	}
}

func TestAnomalySink_ProcessEvents(t *testing.T) {
	detector := NewAnomalyDetector(10*time.Second, 2, 30*time.Second, 1*time.Millisecond)
	sink := NewAnomalySink(50)
	now := time.Now()

	events := []ChangeEvent{
		makeAnomalyEvent(7070, "tcp"),
		makeAnomalyEvent(7070, "tcp"),
		makeAnomalyEvent(7070, "tcp"),
	}

	for i := range events {
		now = now.Add(100 * time.Millisecond)
		sink.ProcessEvents(detector, events[i:i+1], now)
	}

	if sink.Len() == 0 {
		t.Fatal("expected at least one anomaly")
	}
	out := sink.Drain()
	if out[0].Port != 7070 {
		t.Fatalf("unexpected port %d", out[0].Port)
	}
}
