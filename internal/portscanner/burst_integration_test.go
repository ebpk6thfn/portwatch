package portscanner

import (
	"testing"
	"time"
)

// TestBurstAlert_PipelineStyleBurstDetection simulates a rapid stream of
// port-open events and verifies that exactly one burst alert fires and
// subsequent events are suppressed by the cooldown.
func TestBurstAlert_PipelineStyleBurstDetection(t *testing.T) {
	const threshold = 4
	ba := NewBurstAlert(threshold, 30*time.Second, 60*time.Second, "tcp")

	events := make([]ChangeEvent, 10)
	for i := range events {
		events[i] = ChangeEvent{
			Entry: Entry{Protocol: "tcp", Port: uint16(3000 + i)},
			Type:  EventOpened,
		}
	}

	var alerts []*ChangeEvent
	for _, e := range events {
		if a := ba.Observe(e); a != nil {
			alerts = append(alerts, a)
		}
	}

	if len(alerts) != 1 {
		t.Errorf("expected exactly 1 burst alert, got %d", len(alerts))
	}
	if alerts[0].Entry.Process != "[burst-alert]" {
		t.Errorf("unexpected process field: %s", alerts[0].Entry.Process)
	}
}
