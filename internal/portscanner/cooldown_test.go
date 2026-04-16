package portscanner

import (
	"testing"
	"time"
)

func makeCooldownNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestCooldown_FirstCallAllowed(t *testing.T) {
	c := NewCooldown(5 * time.Second)
	if !c.Allow("key1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestCooldown_SecondCallWithinPeriod_Suppressed(t *testing.T) {
	now := time.Now()
	c := NewCooldown(10 * time.Second)
	c.nowFn = makeCooldownNow(now)
	c.Allow("key1")
	if c.Allow("key1") {
		t.Fatal("expected second call within cooldown to be suppressed")
	}
}

func TestCooldown_AfterPeriodExpires_Allowed(t *testing.T) {
	now := time.Now()
	c := NewCooldown(5 * time.Second)
	c.nowFn = makeCooldownNow(now)
	c.Allow("key1")
	c.nowFn = makeCooldownNow(now.Add(6 * time.Second))
	if !c.Allow("key1") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestCooldown_IndependentKeys(t *testing.T) {
	c := NewCooldown(10 * time.Second)
	c.Allow("a")
	if !c.Allow("b") {
		t.Fatal("expected independent key to be allowed")
	}
}

func TestCooldown_Reset_AllowsImmediately(t *testing.T) {
	now := time.Now()
	c := NewCooldown(10 * time.Second)
	c.nowFn = makeCooldownNow(now)
	c.Allow("key1")
	c.Reset("key1")
	if !c.Allow("key1") {
		t.Fatal("expected allow after reset")
	}
}

func TestCooldown_Flush_RemovesExpired(t *testing.T) {
	now := time.Now()
	c := NewCooldown(5 * time.Second)
	c.nowFn = makeCooldownNow(now)
	c.Allow("a")
	c.Allow("b")
	c.nowFn = makeCooldownNow(now.Add(10 * time.Second))
	c.Flush()
	if c.Len() != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", c.Len())
	}
}

func TestCooldown_Len_Tracking(t *testing.T) {
	c := NewCooldown(10 * time.Second)
	c.Allow("x")
	c.Allow("y")
	if c.Len() != 2 {
		t.Fatalf("expected len 2, got %d", c.Len())
	}
}
