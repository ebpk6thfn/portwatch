package portscanner

import (
	"testing"
	"time"
)

func makeBurstNow(base time.Time) func() time.Time {
	current := base
	return func() time.Time { return current }
}

func TestBurstDetector_BelowThreshold(t *testing.T) {
	b := NewBurstDetector(3, 10*time.Second)
	for i := 0; i < 3; i++ {
		if b.Record() {
			t.Errorf("expected no burst on event %d", i+1)
		}
	}
}

func TestBurstDetector_ExceedsThreshold(t *testing.T) {
	b := NewBurstDetector(3, 10*time.Second)
	for i := 0; i < 3; i++ {
		b.Record()
	}
	if !b.Record() {
		t.Error("expected burst on 4th event")
	}
}

func TestBurstDetector_EvictsOldEvents(t *testing.T) {
	base := time.Now()
	b := NewBurstDetector(2, 5*time.Second)

	// record 2 events at t=0
	b.now = func() time.Time { return base }
	b.Record()
	b.Record()

	// advance past window
	b.now = func() time.Time { return base.Add(6 * time.Second) }

	// old events evicted; only this new one counts
	if b.Record() {
		t.Error("expected no burst after window reset")
	}
}

func TestBurstDetector_Count(t *testing.T) {
	b := NewBurstDetector(10, 10*time.Second)
	b.Record()
	b.Record()
	b.Record()
	if got := b.Count(); got != 3 {
		t.Errorf("expected count 3, got %d", got)
	}
}

func TestBurstDetector_Reset(t *testing.T) {
	b := NewBurstDetector(10, 10*time.Second)
	b.Record()
	b.Record()
	b.Reset()
	if got := b.Count(); got != 0 {
		t.Errorf("expected count 0 after reset, got %d", got)
	}
}

func TestBurstDetector_ThresholdOne(t *testing.T) {
	b := NewBurstDetector(1, 10*time.Second)
	if b.Record() {
		t.Error("first event should not trigger burst with threshold 1")
	}
	if !b.Record() {
		t.Error("second event should trigger burst with threshold 1")
	}
}
