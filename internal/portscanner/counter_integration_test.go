package portscanner

import (
	"testing"
	"time"
)

// TestCounter_PipelineStyleBurstDetection simulates a pipeline consumer
// that uses Counter to detect bursty port churn and suppress after threshold.
func TestCounter_PipelineStyleBurstDetection(t *testing.T) {
	const burstThreshold = 3
	window := 5 * time.Second
	base := time.Now()

	c := NewCounter(window)
	c.now = func() time.Time { return base }

	events := []ChangeEvent{
		{Entry: Entry{Port: 9000, Protocol: "tcp"}, Kind: "opened"},
		{Entry: Entry{Port: 9000, Protocol: "tcp"}, Kind: "closed"},
		{Entry: Entry{Port: 9000, Protocol: "tcp"}, Kind: "opened"},
		{Entry: Entry{Port: 9000, Protocol: "tcp"}, Kind: "closed"},
	}

	suppressed := 0
	for _, ev := range events {
		key := ev.Entry.Key()
		n := c.Add(key)
		if n > burstThreshold {
			suppressed++
		}
	}

	// first 3 pass, 4th is suppressed
	if suppressed != 1 {
		t.Fatalf("expected 1 suppressed event, got %d", suppressed)
	}

	// after window expires counts reset naturally
	c.now = func() time.Time { return base.Add(10 * time.Second) }
	if n := c.Count("tcp:9000"); n != 0 {
		t.Fatalf("expected counter to expire, got %d", n)
	}
}
