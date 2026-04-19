package portscanner

import (
	"testing"
	"time"
)

func makeQuotaNow(base time.Time) func() time.Time {
	current := base
	return func() time.Time { return current }
}

func TestQuota_AllowsWithinLimit(t *testing.T) {
	policy := QuotaPolicy{Window: time.Minute, MaxHigh: 3, MaxMedium: 3, MaxLow: 3}
	now := makeQuotaNow(time.Now())
	q := NewQuota(policy, now)

	for i := 0; i < 3; i++ {
		if !q.Allow(SeverityHigh) {
			t.Fatalf("expected allow on call %d", i)
		}
	}
}

func TestQuota_BlocksAtLimit(t *testing.T) {
	policy := QuotaPolicy{Window: time.Minute, MaxHigh: 2, MaxMedium: 10, MaxLow: 10}
	now := makeQuotaNow(time.Now())
	q := NewQuota(policy, now)

	q.Allow(SeverityHigh)
	q.Allow(SeverityHigh)

	if q.Allow(SeverityHigh) {
		t.Fatal("expected block after quota exhausted")
	}
}

func TestQuota_IndependentSeverities(t *testing.T) {
	policy := QuotaPolicy{Window: time.Minute, MaxHigh: 1, MaxMedium: 5, MaxLow: 5}
	now := makeQuotaNow(time.Now())
	q := NewQuota(policy, now)

	q.Allow(SeverityHigh)
	if q.Allow(SeverityHigh) {
		t.Fatal("high should be blocked")
	}
	if !q.Allow(SeverityMedium) {
		t.Fatal("medium should still be allowed")
	}
}

func TestQuota_EvictsExpired(t *testing.T) {
	base := time.Now()
	current := base
	now := func() time.Time { return current }

	policy := QuotaPolicy{Window: time.Minute, MaxHigh: 2, MaxMedium: 10, MaxLow: 10}
	q := NewQuota(policy, now)

	q.Allow(SeverityHigh)
	q.Allow(SeverityHigh)

	// advance past window
	current = base.Add(2 * time.Minute)

	if !q.Allow(SeverityHigh) {
		t.Fatal("expected allow after window expired")
	}
}

func TestQuota_Count(t *testing.T) {
	policy := QuotaPolicy{Window: time.Minute, MaxHigh: 10, MaxMedium: 10, MaxLow: 10}
	now := makeQuotaNow(time.Now())
	q := NewQuota(policy, now)

	q.Allow(SeverityLow)
	q.Allow(SeverityLow)

	if got := q.Count(SeverityLow); got != 2 {
		t.Fatalf("expected count 2, got %d", got)
	}
}

func TestDefaultQuotaPolicy_Valid(t *testing.T) {
	p := DefaultQuotaPolicy()
	if p.MaxHigh <= 0 || p.MaxMedium <= 0 || p.MaxLow <= 0 {
		t.Fatal("default policy has non-positive limits")
	}
	if p.Window <= 0 {
		t.Fatal("default policy has non-positive window")
	}
}
