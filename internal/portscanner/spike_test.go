package portscanner

import (
	"testing"
	"time"
)

func makeSpikeNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func makeSpikeEvents(n int) []ChangeEvent {
	events := make([]ChangeEvent, n)
	for i := range events {
		events[i] = ChangeEvent{Type: EventOpened}
	}
	return events
}

func TestSpikeDetector_BelowThreshold_NoAlert(t *testing.T) {
	policy := SpikePolicy{Window: 10 * time.Second, Threshold: 5, Cooldown: 30 * time.Second}
	d := NewSpikeDetector(policy)

	alert := d.Record(makeSpikeEvents(4))
	if alert != nil {
		t.Fatalf("expected no alert below threshold, got %v", alert)
	}
}

func TestSpikeDetector_AtThreshold_ReturnsAlert(t *testing.T) {
	policy := SpikePolicy{Window: 10 * time.Second, Threshold: 5, Cooldown: 30 * time.Second}
	d := NewSpikeDetector(policy)

	alert := d.Record(makeSpikeEvents(5))
	if alert == nil {
		t.Fatal("expected spike alert at threshold")
	}
	if alert.Count != 5 {
		t.Errorf("expected count 5, got %d", alert.Count)
	}
	if alert.Threshold != 5 {
		t.Errorf("expected threshold 5, got %d", alert.Threshold)
	}
}

func TestSpikeDetector_CooldownSuppressesRepeat(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	policy := SpikePolicy{Window: 10 * time.Second, Threshold: 3, Cooldown: 60 * time.Second}
	d := NewSpikeDetector(policy)
	d.now = makeSpikeNow(base)

	// First spike.
	alert := d.Record(makeSpikeEvents(3))
	if alert == nil {
		t.Fatal("expected first spike alert")
	}

	// Immediately after — still within cooldown.
	alert = d.Record(makeSpikeEvents(3))
	if alert != nil {
		t.Fatalf("expected cooldown to suppress alert, got %v", alert)
	}
}

func TestSpikeDetector_AlertAfterCooldown(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	policy := SpikePolicy{Window: 30 * time.Second, Threshold: 3, Cooldown: 60 * time.Second}
	d := NewSpikeDetector(policy)

	current := base
	d.now = func() time.Time { return current }

	d.Record(makeSpikeEvents(3)) // triggers alert, sets lastAlert

	current = base.Add(90 * time.Second) // past cooldown
	alert := d.Record(makeSpikeEvents(3))
	if alert == nil {
		t.Fatal("expected alert after cooldown expired")
	}
}

func TestSpikeDetector_EvictsOldEvents(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	policy := SpikePolicy{Window: 10 * time.Second, Threshold: 5, Cooldown: 0}
	d := NewSpikeDetector(policy)

	current := base
	d.now = func() time.Time { return current }

	d.Record(makeSpikeEvents(4)) // 4 events at t=0

	current = base.Add(15 * time.Second) // old events now outside window
	alert := d.Record(makeSpikeEvents(1)) // only 1 new event visible
	if alert != nil {
		t.Fatalf("expected no alert after eviction, got %v", alert)
	}
	if d.Count() != 1 {
		t.Errorf("expected count 1 after eviction, got %d", d.Count())
	}
}

func TestSpikeDetector_String(t *testing.T) {
	a := SpikeAlert{Count: 7, Window: 30 * time.Second, Threshold: 5,
		DetectedAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)}
	s := a.String()
	if s == "" {
		t.Error("expected non-empty string from SpikeAlert")
	}
}
