package portscanner

import (
	"testing"
	"time"
)

// TestReaper_PipelineStyleUsage simulates the reaper being used
// alongside a scanner pipeline to evict stale port-tracking entries.
func TestReaper_PipelineStyleUsage(t *testing.T) {
	base := time.Now()
	current := base

	r := NewReaper(ReaperPolicy{MaxAge: 30 * time.Second, Interval: 10 * time.Second})
	r.now = func() time.Time { return current }

	// Simulate scanning: ports observed at t=0
	ports := []string{"tcp:22", "tcp:80", "tcp:443"}
	for _, p := range ports {
		r.Touch(p)
	}

	if r.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", r.Len())
	}

	// Advance time: tcp:22 still active, others gone
	current = base.Add(20 * time.Second)
	r.Touch("tcp:22")

	// Advance past MaxAge for tcp:80 and tcp:443
	current = base.Add(35 * time.Second)
	reaped := r.Reap()

	if len(reaped) != 2 {
		t.Errorf("expected 2 reaped entries, got %d: %v", len(reaped), reaped)
	}

	_, ok := r.Get("tcp:22")
	if !ok {
		t.Error("expected tcp:22 to still be tracked")
	}
}
