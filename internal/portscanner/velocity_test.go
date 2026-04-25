package portscanner

import (
	"testing"
	"time"
)

func makeVelocityNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestVelocity_InitialRate_IsZero(t *testing.T) {
	now := makeVelocityNow(time.Now())
	v := NewVelocity(DefaultVelocityPolicy(), now)
	if r := v.Rate("tcp:80"); r != 0 {
		t.Fatalf("expected 0, got %f", r)
	}
}

func TestVelocity_SingleRecord_ReturnsRate(t *testing.T) {
	base := time.Now()
	now := makeVelocityNow(base)
	v := NewVelocity(DefaultVelocityPolicy(), now)
	rate := v.Record("tcp:80")
	if rate <= 0 {
		t.Fatalf("expected positive rate, got %f", rate)
	}
}

func TestVelocity_MultipleRecords_AccumulateRate(t *testing.T) {
	base := time.Now()
	now := makeVelocityNow(base)
	v := NewVelocity(DefaultVelocityPolicy(), now)

	for i := 0; i < 5; i++ {
		v.Record("tcp:443")
	}
	rate := v.Rate("tcp:443")
	if rate <= 0 {
		t.Fatalf("expected positive rate after 5 records, got %f", rate)
	}
}

func TestVelocity_EvictsOldEntries(t *testing.T) {
	base := time.Now()
	current := base
	nowFn := func() time.Time { return current }

	policy := VelocityPolicy{Window: 30 * time.Second, MaxItems: 1000}
	v := NewVelocity(policy, nowFn)

	v.Record("tcp:22")
	v.Record("tcp:22")

	// Advance past the window
	current = base.Add(60 * time.Second)

	rate := v.Rate("tcp:22")
	if rate != 0 {
		t.Fatalf("expected 0 after eviction, got %f", rate)
	}
}

func TestVelocity_IndependentKeys(t *testing.T) {
	base := time.Now()
	now := makeVelocityNow(base)
	v := NewVelocity(DefaultVelocityPolicy(), now)

	v.Record("tcp:80")
	v.Record("tcp:80")
	v.Record("tcp:80")

	v.Record("udp:53")

	rate80 := v.Rate("tcp:80")
	rate53 := v.Rate("udp:53")

	if rate80 <= rate53 {
		t.Fatalf("expected tcp:80 rate (%f) > udp:53 rate (%f)", rate80, rate53)
	}
}

func TestVelocity_String_ContainsKey(t *testing.T) {
	now := makeVelocityNow(time.Now())
	v := NewVelocity(DefaultVelocityPolicy(), now)
	v.Record("tcp:8080")
	s := v.String("tcp:8080")
	if len(s) == 0 {
		t.Fatal("expected non-empty string")
	}
	if s[:8] != "velocity" {
		t.Fatalf("unexpected string prefix: %s", s)
	}
}

func TestVelocity_MaxItems_Enforced(t *testing.T) {
	now := makeVelocityNow(time.Now())
	policy := VelocityPolicy{Window: time.Hour, MaxItems: 5}
	v := NewVelocity(policy, now)

	for i := 0; i < 20; i++ {
		v.Record("tcp:9090")
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	if len(v.buckets["tcp:9090"]) > 5 {
		t.Fatalf("expected max 5 entries, got %d", len(v.buckets["tcp:9090"]))
	}
}
