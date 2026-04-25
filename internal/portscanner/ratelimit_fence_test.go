package portscanner

import (
	"testing"
	"time"
)

func fixedFenceNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestFence_BelowThreshold_Allows(t *testing.T) {
	policy := FencePolicy{
		MaxEvents:          5,
		Window:             time.Minute,
		CooldownAfterFence: 30 * time.Second,
	}
	f := NewFence(policy)

	for i := 0; i < 5; i++ {
		if !f.Allow() {
			t.Fatalf("expected allow on event %d", i+1)
		}
	}
}

func TestFence_AtThreshold_Trips(t *testing.T) {
	policy := FencePolicy{
		MaxEvents:          3,
		Window:             time.Minute,
		CooldownAfterFence: time.Minute,
	}
	f := NewFence(policy)

	// First 3 allowed.
	for i := 0; i < 3; i++ {
		f.Allow()
	}
	// 4th should trip the fence.
	if f.Allow() {
		t.Fatal("expected fence to block 4th event")
	}
	if !f.IsFenced() {
		t.Fatal("expected IsFenced to be true")
	}
}

func TestFence_CooldownLifts(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	policy := FencePolicy{
		MaxEvents:          2,
		Window:             time.Minute,
		CooldownAfterFence: 10 * time.Second,
	}
	f := NewFence(policy)
	f.now = fixedFenceNow(base)

	// Trip the fence.
	for i := 0; i < 3; i++ {
		f.Allow()
	}
	if !f.IsFenced() {
		t.Fatal("expected fence to be active")
	}

	// Advance past cooldown.
	f.now = fixedFenceNow(base.Add(11 * time.Second))
	if !f.Allow() {
		t.Fatal("expected fence to lift after cooldown")
	}
	if f.IsFenced() {
		t.Fatal("expected IsFenced to be false after cooldown")
	}
}

func TestFence_EvictsOldEvents(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	policy := FencePolicy{
		MaxEvents:          3,
		Window:             time.Minute,
		CooldownAfterFence: time.Minute,
	}
	f := NewFence(policy)
	f.now = fixedFenceNow(base)

	// Push 3 events at base time.
	for i := 0; i < 3; i++ {
		f.Allow()
	}

	// Advance past window — old events should be evicted.
	f.now = fixedFenceNow(base.Add(61 * time.Second))
	if !f.Allow() {
		t.Fatal("expected allow after old events evicted")
	}
}

func TestFence_Reset_ClearsState(t *testing.T) {
	policy := FencePolicy{
		MaxEvents:          1,
		Window:             time.Minute,
		CooldownAfterFence: time.Hour,
	}
	f := NewFence(policy)

	f.Allow()
	f.Allow() // trips fence

	if !f.IsFenced() {
		t.Fatal("expected fence to be active before reset")
	}

	f.Reset()

	if f.IsFenced() {
		t.Fatal("expected fence to be cleared after reset")
	}
	if !f.Allow() {
		t.Fatal("expected allow after reset")
	}
}

func TestFence_BlocksDuringCooldown(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	policy := FencePolicy{
		MaxEvents:          1,
		Window:             time.Minute,
		CooldownAfterFence: 30 * time.Second,
	}
	f := NewFence(policy)
	f.now = fixedFenceNow(base)

	f.Allow()
	f.Allow() // trips

	// Advance but still within cooldown.
	f.now = fixedFenceNow(base.Add(15 * time.Second))
	if f.Allow() {
		t.Fatal("expected block during cooldown")
	}
}
