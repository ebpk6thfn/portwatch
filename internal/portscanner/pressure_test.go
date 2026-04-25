package portscanner

import (
	"testing"
	"time"
)

func fixedPressureNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestPressureGauge_InitialLevel_IsLow(t *testing.T) {
	g := NewPressureGauge(DefaultPressurePolicy())
	if got := g.Level(); got != PressureLow {
		t.Fatalf("expected Low, got %s", got)
	}
}

func TestPressureGauge_BelowLowWatermark_IsLow(t *testing.T) {
	policy := PressurePolicy{LowWatermark: 10, HighWatermark: 50, Window: time.Minute}
	g := NewPressureGauge(policy)
	g.Record(5)
	g.Record(3)
	if got := g.Level(); got != PressureLow {
		t.Fatalf("expected Low, got %s", got)
	}
}

func TestPressureGauge_AboveHighWatermark_IsHigh(t *testing.T) {
	policy := PressurePolicy{LowWatermark: 10, HighWatermark: 50, Window: time.Minute}
	g := NewPressureGauge(policy)
	g.Record(60)
	g.Record(70)
	if got := g.Level(); got != PressureHigh {
		t.Fatalf("expected High, got %s", got)
	}
}

func TestPressureGauge_BetweenWatermarks_IsMedium(t *testing.T) {
	policy := PressurePolicy{LowWatermark: 10, HighWatermark: 50, Window: time.Minute}
	g := NewPressureGauge(policy)
	g.Record(20)
	g.Record(30)
	if got := g.Level(); got != PressureMedium {
		t.Fatalf("expected Medium, got %s", got)
	}
}

func TestPressureGauge_EvictsOldSamples(t *testing.T) {
	base := time.Now()
	policy := PressurePolicy{LowWatermark: 10, HighWatermark: 50, Window: 10 * time.Second}
	g := NewPressureGauge(policy)
	g.now = fixedPressureNow(base)
	g.Record(80) // high pressure sample, old

	g.now = fixedPressureNow(base.Add(15 * time.Second))
	g.Record(2) // low pressure sample, recent

	if got := g.Level(); got != PressureLow {
		t.Fatalf("expected Low after eviction, got %s", got)
	}
}

func TestPressureGauge_Depth_TracksRetained(t *testing.T) {
	base := time.Now()
	policy := PressurePolicy{LowWatermark: 10, HighWatermark: 50, Window: 10 * time.Second}
	g := NewPressureGauge(policy)
	g.now = fixedPressureNow(base)
	g.Record(5)
	g.Record(5)
	if g.Depth() != 2 {
		t.Fatalf("expected depth 2, got %d", g.Depth())
	}
	g.now = fixedPressureNow(base.Add(20 * time.Second))
	g.Record(5)
	if g.Depth() != 1 {
		t.Fatalf("expected depth 1 after eviction, got %d", g.Depth())
	}
}

func TestPressureLevel_String(t *testing.T) {
	cases := []struct {
		level PressureLevel
		want  string
	}{
		{PressureLow, "low"},
		{PressureMedium, "medium"},
		{PressureHigh, "high"},
	}
	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("PressureLevel(%d).String() = %q, want %q", tc.level, got, tc.want)
		}
	}
}
