package portscanner

import (
	"fmt"
	"testing"
	"time"
)

// TestWindowCounter_BurstDetection simulates pipeline-style burst detection
// where events for a port are counted over a sliding window.
func TestWindowCounter_BurstDetection(t *testing.T) {
	base := time.Now()
	wc := NewWindowCounter(30 * time.Second)
	wc.nowFn = func() time.Time { return base }

	port := "tcp:8080"
	threshold := 5

	// Simulate rapid open/close events
	for i := 0; i < threshold; i++ {
		wc.nowFn = func() time.Time { return base.Add(time.Duration(i) * time.Second) }
		count := wc.Add(port)
		if count > threshold {
			t.Fatalf("unexpected burst at step %d: count=%d", i, count)
		}
	}

	// One more should hit threshold
	wc.nowFn = func() time.Time { return base.Add(5 * time.Second) }
	count := wc.Add(port)
	if count != threshold+1 {
		t.Fatalf("expected burst count %d, got %d", threshold+1, count)
	}

	// After window expires, count resets naturally
	wc.nowFn = func() time.Time { return base.Add(60 * time.Second) }
	wc.Add(port)
	after := wc.Count(port)
	if after != 1 {
		t.Fatalf("expected count=1 after window expiry, got %d", after)
	}
	fmt.Printf("burst detection ok, post-window count=%d\n", after)
}
