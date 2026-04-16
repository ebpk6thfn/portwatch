package portscanner

import (
	"testing"
	"time"
)

// TestTrend_PipelineStyleBurstDetection simulates recording scan results
// over time and verifying trend detection in a pipeline-like scenario.
func TestTrend_PipelineStyleBurstDetection(t *testing.T) {
	tr := NewTrendTracker(2 * time.Minute)
	base := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	// Simulate quiet period
	for i := 0; i < 5; i++ {
		tr.Record(base.Add(time.Duration(i)*10*time.Second), 1)
	}

	// Simulate burst
	burst := base.Add(60 * time.Second)
	tr.Record(burst, 12)

	if got := tr.Trend(burst); got != TrendUp {
		t.Fatalf("expected burst to produce TrendUp, got %s", got)
	}

	// Simulate recovery
	recovery := burst.Add(30 * time.Second)
	tr.Record(recovery, 1)
	if got := tr.Trend(recovery); got != TrendDown {
		t.Fatalf("expected recovery to produce TrendDown, got %s", got)
	}
}
