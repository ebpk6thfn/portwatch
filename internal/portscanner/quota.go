package portscanner

import (
	"sync"
	"time"
)

// QuotaPolicy defines per-severity alert quotas within a rolling window.
type QuotaPolicy struct {
	Window   time.Duration
	MaxHigh  int
	MaxMedium int
	MaxLow   int
}

// DefaultQuotaPolicy returns sensible defaults.
func DefaultQuotaPolicy() QuotaPolicy {
	return QuotaPolicy{
		Window:    time.Hour,
		MaxHigh:   50,
		MaxMedium: 100,
		MaxLow:    200,
	}
}

// Quota enforces per-severity event quotas over a rolling time window.
type Quota struct {
	mu     sync.Mutex
	policy QuotaPolicy
	now    func() time.Time
	buckets map[Severity][]time.Time
}

// NewQuota creates a Quota enforcer with the given policy.
func NewQuota(policy QuotaPolicy, now func() time.Time) *Quota {
	if now == nil {
		now = time.Now
	}
	return &Quota{
		policy:  policy,
		now:     now,
		buckets: make(map[Severity][]time.Time),
	}
}

// Allow returns true if the event is within quota, recording it if so.
func (q *Quota) Allow(sev Severity) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	cutoff := now.Add(-q.policy.Window)

	// evict expired
	times := q.buckets[sev]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	q.buckets[sev] = valid

	max := q.maxFor(sev)
	if max >= 0 && len(valid) >= max {
		return false
	}
	q.buckets[sev] = append(q.buckets[sev], now)
	return true
}

// Count returns the current count for a severity within the window.
func (q *Quota) Count(sev Severity) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.buckets[sev])
}

func (q *Quota) maxFor(sev Severity) int {
	switch sev {
	case SeverityHigh:
		return q.policy.MaxHigh
	case SeverityMedium:
		return q.policy.MaxMedium
	case SeverityLow:
		return q.policy.MaxLow
	}
	return -1
}
