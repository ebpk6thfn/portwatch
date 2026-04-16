package portscanner

import (
	"testing"
	"time"
)

func makeDecayNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestDecayCounter_EmptyScore(t *testing.T) {
	dc := NewDecayCounter(10 * time.Second)
	if dc.Score() != 0 {
		t.Fatalf("expected 0 score on empty counter")
	}
}

func TestDecayCounter_SingleEvent_ScoreNearOne(t *testing.T) {
	base := time.Now()
	dc := NewDecayCounter(10 * time.Second)
	dc.now = func() time.Time { return base }
	dc.Add()
	score := dc.Score()
	if score <= 0.99 || score > 1.0 {
		t.Fatalf("expected score ~1.0 for immediate event, got %f", score)
	}
}

func TestDecayCounter_EventHalfwayThrough_ScoreHalf(t *testing.T) {
	base := time.Now()
	dc := NewDecayCounter(10 * time.Second)
	dc.now = func() time.Time { return base }
	dc.Add()
	// advance 5s (halfway through window)
	dc.now = func() time.Time { return base.Add(5 * time.Second) }
	score := dc.Score()
	if score < 0.49 || score > 0.51 {
		t.Fatalf("expected score ~0.5, got %f", score)
	}
}

func TestDecayCounter_EventExpires(t *testing.T) {
	base := time.Now()
	dc := NewDecayCounter(10 * time.Second)
	dc.now = func() time.Time { return base }
	dc.Add()
	dc.now = func() time.Time { return base.Add(11 * time.Second) }
	if dc.Count() != 0 {
		t.Fatal("expected event to be evicted after window")
	}
	if dc.Score() != 0 {
		t.Fatal("expected score 0 after eviction")
	}
}

func TestDecayCounter_MultipleEvents_ScoreAccumulates(t *testing.T) {
	base := time.Now()
	dc := NewDecayCounter(10 * time.Second)
	dc.now = func() time.Time { return base }
	dc.Add()
	dc.Add()
	dc.Add()
	score := dc.Score()
	if score < 2.9 {
		t.Fatalf("expected score ~3 for 3 immediate events, got %f", score)
	}
}

func TestDecayCounter_Count_DoesNotDecay(t *testing.T) {
	base := time.Now()
	dc := NewDecayCounter(10 * time.Second)
	dc.now = func() time.Time { return base }
	dc.Add()
	dc.Add()
	if dc.Count() != 2 {
		t.Fatalf("expected count 2, got %d", dc.Count())
	}
}
