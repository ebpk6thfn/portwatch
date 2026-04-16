package portscanner

import (
	"testing"
	"time"
)

// TestCooldown_PipelineStyleUsage simulates using Cooldown as an event gate
// in a pipeline: events for the same key are suppressed until the window expires.
func TestCooldown_PipelineStyleUsage(t *testing.T) {
	base := time.Now()
	c := NewCooldown(3 * time.Second)
	c.nowFn = func() time.Time { return base }

	events := []struct {
		offset  time.Duration
		key     string
		wantAllow bool
	}{
		{0, "port:tcp:80", true},
		{1 * time.Second, "port:tcp:80", false},
		{2 * time.Second, "port:tcp:443", true},
		{3 * time.Second, "port:tcp:80", false}, // exactly at boundary, not yet expired
		{4 * time.Second, "port:tcp:80", true},  // past the 3s window
		{4 * time.Second, "port:tcp:443", false},
	}

	for _, ev := range events {
		c.nowFn = func(o time.Duration) func() time.Time {
			return func() time.Time { return base.Add(o) }
		}(ev.offset)
		got := c.Allow(ev.key)
		if got != ev.wantAllow {
			t.Errorf("offset=%v key=%s: Allow()=%v, want %v",
				ev.offset, ev.key, got, ev.wantAllow)
		}
	}
}
