package portscanner

import (
	"testing"
	"time"
)

// TestCounter_PipelineStyleBurstDetection simulates a pipeline consumer
// that uses Counter to detect bursty port churn and suppress after threshold.
// It verifies that events exceeding burstThreshold within the window are
// suppressed, and that counts reset once the window expires.
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

// TestCounter_MultipleKeysBurstIsolation verifies that burst suppression
// is tracked independently per key and does not bleed across ports.
func TestCounter_MultipleKeysBurstIsolation(t *testing.T) {
	const burstThreshold = 2
	window := 5 * time.Second
	base := time.Now()

	c := NewCounter(window)
	c.now = func() time.Time { return base }

	keys := []string{"tcp:8080", "tcp:8081", "tcp:8080", "tcp:8081", "tcp:8080"}
	counts := map[string]int{}
	for _, key := range keys {
		counts[key] = c.Add(key)
	}

	// tcp:8080 was added 3 times, tcp:8081 twice
	if counts["tcp:8080"] != 3 {
		t.Fatalf("expected tcp:8080 count 3, got %d", counts["tcp:8080"])
	}
	if counts["tcp:8081"] != 2 {
		t.Fatalf("expected tcp:8081 count 2, got %d", counts["tcp:8081"])
	}
}
