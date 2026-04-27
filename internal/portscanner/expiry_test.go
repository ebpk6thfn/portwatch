package portscanner

import (
	"testing"
	"time"
)

func makeExpiryNow(base time.Time) func() time.Time {
	current := base
	return func() time.Time { return current }
}

func advanceExpiry(nowFn *func() time.Time, base *time.Time, d time.Duration) {
	*base = base.Add(d)
	*nowFn = func() time.Time { return *base }
}

func TestExpiryTracker_TouchCreatesEntry(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	nowFn := func() time.Time { return base }
	tr := NewExpiryTracker(DefaultExpiryPolicy(), nowFn)

	e := tr.Touch("tcp:8080")
	if e.Key != "tcp:8080" {
		t.Fatalf("expected key tcp:8080, got %s", e.Key)
	}
	if !e.FirstSeen.Equal(base) {
		t.Fatalf("unexpected FirstSeen")
	}
	if tr.Len() != 1 {
		t.Fatalf("expected len 1, got %d", tr.Len())
	}
}

func TestExpiryTracker_TouchUpdatesLastSeen(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	nowFn := func() time.Time { return base }
	tr := NewExpiryTracker(DefaultExpiryPolicy(), nowFn)

	tr.Touch("tcp:9090")

	base = base.Add(30 * time.Second)
	nowFn = func() time.Time { return base }
	tr.now = nowFn

	e := tr.Touch("tcp:9090")
	if !e.LastSeen.Equal(base) {
		t.Fatalf("LastSeen not updated")
	}
	if tr.Len() != 1 {
		t.Fatalf("expected len 1, got %d", tr.Len())
	}
}

func TestExpiryTracker_NoExpiredBeforeTTL(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	nowFn := func() time.Time { return base }
	policy := ExpiryPolicy{TTL: 2 * time.Minute}
	tr := NewExpiryTracker(policy, nowFn)

	tr.Touch("udp:53")

	base = base.Add(90 * time.Second)
	tr.now = func() time.Time { return base }

	expired := tr.Expired()
	if len(expired) != 0 {
		t.Fatalf("expected no expired entries, got %d", len(expired))
	}
}

func TestExpiryTracker_ExpiredAfterTTL(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	nowFn := func() time.Time { return base }
	policy := ExpiryPolicy{TTL: 1 * time.Minute}
	tr := NewExpiryTracker(policy, nowFn)

	tr.Touch("tcp:443")

	base = base.Add(2 * time.Minute)
	tr.now = func() time.Time { return base }

	expired := tr.Expired()
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired entry, got %d", len(expired))
	}
	if expired[0].Key != "tcp:443" {
		t.Fatalf("unexpected key %s", expired[0].Key)
	}
	if expired[0].ExpiredAt.IsZero() {
		t.Fatal("ExpiredAt should be set")
	}
	if tr.Len() != 0 {
		t.Fatalf("expected len 0 after expiry, got %d", tr.Len())
	}
}

func TestExpiryTracker_MultipleKeys_IndependentTTL(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	nowFn := func() time.Time { return base }
	policy := ExpiryPolicy{TTL: 1 * time.Minute}
	tr := NewExpiryTracker(policy, nowFn)

	tr.Touch("tcp:80")

	base = base.Add(45 * time.Second)
	tr.now = func() time.Time { return base }
	tr.Touch("tcp:443") // touched later, should not expire yet

	base = base.Add(20 * time.Second) // 65s after tcp:80 first touch, 20s after tcp:443
	tr.now = func() time.Time { return base }

	expired := tr.Expired()
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired, got %d", len(expired))
	}
	if expired[0].Key != "tcp:80" {
		t.Fatalf("expected tcp:80 to expire, got %s", expired[0].Key)
	}
	if tr.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", tr.Len())
	}
}

func TestExpiryTracker_DefaultPolicy(t *testing.T) {
	p := DefaultExpiryPolicy()
	if p.TTL <= 0 {
		t.Fatal("default TTL should be positive")
	}
}
