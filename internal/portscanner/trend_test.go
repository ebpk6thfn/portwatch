package portscanner

import (
	"testing"
	"time"
)

var trendBase = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestTrend_StableWithNoPoints(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	if got := tr.Trend(trendBase); got != TrendStable {
		t.Fatalf("expected stable, got %s", got)
	}
}

func TestTrend_StableWithOnePoint(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	tr.Record(trendBase, 5)
	if got := tr.Trend(trendBase); got != TrendStable {
		t.Fatalf("expected stable, got %s", got)
	}
}

func TestTrend_Up(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	tr.Record(trendBase, 2)
	tr.Record(trendBase.Add(10*time.Second), 5)
	if got := tr.Trend(trendBase.Add(10 * time.Second)); got != TrendUp {
		t.Fatalf("expected up, got %s", got)
	}
}

func TestTrend_Down(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	tr.Record(trendBase, 10)
	tr.Record(trendBase.Add(10*time.Second), 3)
	if got := tr.Trend(trendBase.Add(10 * time.Second)); got != TrendDown {
		t.Fatalf("expected down, got %s", got)
	}
}

func TestTrend_EvictsOldPoints(t *testing.T) {
	tr := NewTrendTracker(30 * time.Second)
	tr.Record(trendBase, 10)
	tr.Record(trendBase.Add(10*time.Second), 8)
	// advance past window; old points evicted
	now := trendBase.Add(2 * time.Minute)
	tr.Record(now, 1)
	if got := tr.Trend(now); got != TrendStable {
		t.Fatalf("expected stable after eviction, got %s", got)
	}
}

func TestTrend_PointsReturnsWindow(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	tr.Record(trendBase, 1)
	tr.Record(trendBase.Add(5*time.Second), 2)
	pts := tr.Points(trendBase.Add(5 * time.Second))
	if len(pts) != 2 {
		t.Fatalf("expected 2 points, got %d", len(pts))
	}
}

func TestTrend_PointsExcludesExpired(t *testing.T) {
	tr := NewTrendTracker(10 * time.Second)
	tr.Record(trendBase, 5)
	now := trendBase.Add(20 * time.Second)
	tr.Record(now, 3)
	pts := tr.Points(now)
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if pts[0].Count != 3 {
		t.Fatalf("expected count 3, got %d", pts[0].Count)
	}
}
