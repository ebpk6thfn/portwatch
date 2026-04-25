package config

import (
	"testing"
	"time"
)

func TestDefaultFenceConfig_Valid(t *testing.T) {
	cfg := DefaultFenceConfig()
	policy, err := BuildFencePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxEvents != 50 {
		t.Errorf("expected MaxEvents=50, got %d", policy.MaxEvents)
	}
	if policy.Window != time.Minute {
		t.Errorf("expected Window=1m, got %v", policy.Window)
	}
	if policy.CooldownAfterFence != 2*time.Minute {
		t.Errorf("expected CooldownAfterFence=2m, got %v", policy.CooldownAfterFence)
	}
}

func TestBuildFencePolicy_CustomValues(t *testing.T) {
	cfg := FenceConfig{
		MaxEvents:          100,
		Window:             "30s",
		CooldownAfterFence: "5m",
	}
	policy, err := BuildFencePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxEvents != 100 {
		t.Errorf("expected 100, got %d", policy.MaxEvents)
	}
	if policy.Window != 30*time.Second {
		t.Errorf("expected 30s, got %v", policy.Window)
	}
	if policy.CooldownAfterFence != 5*time.Minute {
		t.Errorf("expected 5m, got %v", policy.CooldownAfterFence)
	}
}

func TestBuildFencePolicy_ZeroMaxEvents_Error(t *testing.T) {
	cfg := FenceConfig{MaxEvents: 0, Window: "1m", CooldownAfterFence: "1m"}
	_, err := BuildFencePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for zero max_events")
	}
}

func TestBuildFencePolicy_InvalidWindow_Error(t *testing.T) {
	cfg := FenceConfig{MaxEvents: 10, Window: "not-a-duration", CooldownAfterFence: "1m"}
	_, err := BuildFencePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for invalid window")
	}
}

func TestBuildFencePolicy_InvalidCooldown_Error(t *testing.T) {
	cfg := FenceConfig{MaxEvents: 10, Window: "1m", CooldownAfterFence: "bad"}
	_, err := BuildFencePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for invalid cooldown")
	}
}

func TestBuildFencePolicy_NegativeWindow_Error(t *testing.T) {
	cfg := FenceConfig{MaxEvents: 10, Window: "-1m", CooldownAfterFence: "1m"}
	_, err := BuildFencePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}
