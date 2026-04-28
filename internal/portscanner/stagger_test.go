package portscanner

import (
	"testing"
	"time"
)

func TestStagger_FirstKey_GetsSlotZero(t *testing.T) {
	s := newStaggerWithClock(DefaultStaggerPolicy(), time.Now)
	d := s.Delay("tcp:80")
	if d != 0 {
		t.Errorf("first key should get slot 0 (delay 0), got %v", d)
	}
}

func TestStagger_SecondKey_GetsNonZeroDelay(t *testing.T) {
	policy := StaggerPolicy{Window: 10 * time.Second, MaxSlot: 10}
	s := newStaggerWithClock(policy, time.Now)
	_ = s.Delay("tcp:80")
	d := s.Delay("tcp:443")
	if d == 0 {
		t.Errorf("second key should get non-zero delay, got %v", d)
	}
}

func TestStagger_SameKey_SameDelay(t *testing.T) {
	s := newStaggerWithClock(DefaultStaggerPolicy(), time.Now)
	d1 := s.Delay("tcp:8080")
	d2 := s.Delay("tcp:8080")
	if d1 != d2 {
		t.Errorf("same key should produce same delay: got %v and %v", d1, d2)
	}
}

func TestStagger_SlotWrapsAtMax(t *testing.T) {
	policy := StaggerPolicy{Window: 10 * time.Second, MaxSlot: 3}
	s := newStaggerWithClock(policy, time.Now)

	keys := []string{"a", "b", "c", "d"}
	delays := make([]time.Duration, len(keys))
	for i, k := range keys {
		delays[i] = s.Delay(k)
	}
	// slot for "d" wraps to slot 0, same as "a"
	if delays[3] != delays[0] {
		t.Errorf("slot should wrap: key d delay %v should equal key a delay %v", delays[3], delays[0])
	}
}

func TestStagger_DelayWithinWindow(t *testing.T) {
	policy := StaggerPolicy{Window: 10 * time.Second, MaxSlot: 10}
	s := newStaggerWithClock(policy, time.Now)

	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		d := s.Delay(key)
		if d >= policy.Window {
			t.Errorf("delay %v for key %q exceeds window %v", d, key, policy.Window)
		}
	}
}

func TestStagger_Reset_ClearsSlots(t *testing.T) {
	s := newStaggerWithClock(DefaultStaggerPolicy(), time.Now)
	_ = s.Delay("tcp:80")
	_ = s.Delay("tcp:443")
	if s.Len() != 2 {
		t.Fatalf("expected 2 slots, got %d", s.Len())
	}
	s.Reset()
	if s.Len() != 0 {
		t.Errorf("expected 0 slots after reset, got %d", s.Len())
	}
}

func TestStagger_DefaultPolicy_NonZeroWindow(t *testing.T) {
	p := DefaultStaggerPolicy()
	if p.Window <= 0 {
		t.Errorf("default window should be positive, got %v", p.Window)
	}
	if p.MaxSlot <= 0 {
		t.Errorf("default max slot should be positive, got %d", p.MaxSlot)
	}
}
