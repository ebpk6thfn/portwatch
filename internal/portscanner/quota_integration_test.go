package portscanner

import (
	"testing"
	"time"
)

// TestQuota_PipelineStyleUsage simulates the quota enforcer in a pipeline
// where events are filtered based on per-severity quotas.
func TestQuota_PipelineStyleUsage(t *testing.T) {
	base := time.Now()
	current := base
	now := func() time.Time { return current }

	policy := QuotaPolicy{
		Window:    time.Minute,
		MaxHigh:   2,
		MaxMedium: 3,
		MaxLow:    10,
	}
	q := NewQuota(policy, now)

	events := []Severity{
		SeverityHigh, SeverityHigh, SeverityHigh, // 3rd should be blocked
		SeverityMedium, SeverityMedium, SeverityMedium, SeverityMedium, // 4th blocked
	}

	allowed := 0
	blocked := 0
	for _, sev := range events {
		if q.Allow(sev) {
			allowed++
		} else {
			blocked++
		}
	}

	if allowed != 5 {
		t.Errorf("expected 5 allowed, got %d", allowed)
	}
	if blocked != 2 {
		t.Errorf("expected 2 blocked, got %d", blocked)
	}

	// advance window — all quotas reset
	current = base.Add(2 * time.Minute)
	if !q.Allow(SeverityHigh) {
		t.Error("expected high to be allowed after window reset")
	}
}
