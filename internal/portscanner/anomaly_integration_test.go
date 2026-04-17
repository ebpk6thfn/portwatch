package portscanner

import (
	"testing"
	"time"
)

// TestAnomalyDetector_PipelineStyleUsage simulates how the daemon would wire
// AnomalyDetector + AnomalySink into a scan loop.
func TestAnomalyDetector_PipelineStyleUsage(t *testing.T) {
	detector := NewAnomalyDetector(
		5*time.Second,  // burst window
		3,              // burst threshold
		15*time.Second, // decay half-life
		1*time.Millisecond, // cooldown (tiny for test)
	)
	sink := NewAnomalySink(100)

	now := time.Now()
	port := uint16(2222)

	// Simulate 6 rapid open events on same port — should trigger burst.
	for i := 0; i < 6; i++ {
		now = now.Add(200 * time.Millisecond)
		ev := ChangeEvent{
			Entry: Entry{Port: port, Protocol: "tcp", Addr: "0.0.0.0"},
			Type:  EventOpened,
		}
		sink.ProcessEvents(detector, []ChangeEvent{ev}, now)
	}

	anomalies := sink.Drain()
	if len(anomalies) == 0 {
		t.Fatal("expected anomalies from burst activity")
	}
	for _, a := range anomalies {
		if a.Port != port {
			t.Errorf("unexpected port %d in anomaly", a.Port)
		}
		if a.Score <= 0 {
			t.Errorf("expected positive score, got %.2f", a.Score)
		}
	}
}
