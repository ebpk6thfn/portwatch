package portscanner

import (
	"strings"
	"testing"
	"time"
)

func TestBuildReport_EmptyTracker(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	now := time.Now()
	rep := BuildReport(tr, now)
	if rep.Direction != TrendStable {
		t.Fatalf("expected stable, got %s", rep.Direction)
	}
	if rep.PointCount != 0 {
		t.Fatalf("expected 0 points, got %d", rep.PointCount)
	}
}

func TestBuildReport_UpTrend(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tr.Record(base, 1)
	tr.Record(base.Add(10*time.Second), 9)
	rep := BuildReport(tr, base.Add(10*time.Second))
	if rep.Direction != TrendUp {
		t.Fatalf("expected up, got %s", rep.Direction)
	}
	if rep.First != 1 || rep.Last != 9 {
		t.Fatalf("unexpected first/last: %d/%d", rep.First, rep.Last)
	}
}

func TestBuildReport_String_ContainsTrend(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tr.Record(base, 3)
	tr.Record(base.Add(5*time.Second), 1)
	rep := BuildReport(tr, base.Add(5*time.Second))
	s := rep.String()
	if !strings.Contains(s, "trend=down") {
		t.Fatalf("expected trend=down in string, got: %s", s)
	}
	if !strings.Contains(s, "points=2") {
		t.Fatalf("expected points=2 in string, got: %s", s)
	}
}

func TestBuildReport_String_NoFirstLast_WhenEmpty(t *testing.T) {
	tr := NewTrendTracker(time.Minute)
	rep := BuildReport(tr, time.Now())
	s := rep.String()
	if strings.Contains(s, "first=") {
		t.Fatalf("did not expect first= in empty report, got: %s", s)
	}
}
