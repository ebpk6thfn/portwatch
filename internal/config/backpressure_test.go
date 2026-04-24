package config

import (
	"testing"
	"time"
)

func TestDefaultBackpressureConfig_Valid(t *testing.T) {
	cfg := DefaultBackpressureConfig()
	policy, err := BuildBackpressurePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.HighWatermark != 100 {
		t.Errorf("expected HighWatermark 100, got %d", policy.HighWatermark)
	}
	if policy.LowWatermark != 25 {
		t.Errorf("expected LowWatermark 25, got %d", policy.LowWatermark)
	}
	if policy.CooldownPeriod != 5*time.Second {
		t.Errorf("expected CooldownPeriod 5s, got %v", policy.CooldownPeriod)
	}
}

func TestBuildBackpressurePolicy_CustomValues(t *testing.T) {
	cfg := BackpressureConfig{
		HighWatermark:  200,
		LowWatermark:   50,
		CooldownPeriod: "10s",
	}
	policy, err := BuildBackpressurePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.HighWatermark != 200 {
		t.Errorf("expected 200, got %d", policy.HighWatermark)
	}
	if policy.CooldownPeriod != 10*time.Second {
		t.Errorf("expected 10s, got %v", policy.CooldownPeriod)
	}
}

func TestBuildBackpressurePolicy_ZeroHighWatermark_Error(t *testing.T) {
	cfg := BackpressureConfig{HighWatermark: 0, LowWatermark: 0, CooldownPeriod: "1s"}
	_, err := BuildBackpressurePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for zero high_watermark")
	}
}

func TestBuildBackpressurePolicy_LowGeHigh_Error(t *testing.T) {
	cfg := BackpressureConfig{HighWatermark: 10, LowWatermark: 10, CooldownPeriod: "1s"}
	_, err := BuildBackpressurePolicy(cfg)
	if err == nil {
		t.Fatal("expected error when low_watermark >= high_watermark")
	}
}

func TestBuildBackpressurePolicy_InvalidDuration_Error(t *testing.T) {
	cfg := BackpressureConfig{HighWatermark: 100, LowWatermark: 25, CooldownPeriod: "notaduration"}
	_, err := BuildBackpressurePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for invalid duration string")
	}
}

func TestBuildBackpressurePolicy_EmptyCooldown_UsesDefault(t *testing.T) {
	cfg := BackpressureConfig{HighWatermark: 50, LowWatermark: 10, CooldownPeriod: ""}
	policy, err := BuildBackpressurePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.CooldownPeriod != 5*time.Second {
		t.Errorf("expected default cooldown 5s, got %v", policy.CooldownPeriod)
	}
}
