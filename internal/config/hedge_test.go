package config

import (
	"testing"
	"time"
)

func TestDefaultHedgeConfig_Valid(t *testing.T) {
	cfg := DefaultHedgeConfig()
	policy, err := BuildHedgePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error from default config: %v", err)
	}
	if policy.Window <= 0 {
		t.Errorf("expected positive window, got %v", policy.Window)
	}
	if policy.MaxPending <= 0 {
		t.Errorf("expected positive max_pending, got %d", policy.MaxPending)
	}
}

func TestBuildHedgePolicy_CustomValues(t *testing.T) {
	cfg := HedgeConfig{Window: "5s", MaxPending: 128}
	policy, err := BuildHedgePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != 5*time.Second {
		t.Errorf("expected 5s, got %v", policy.Window)
	}
	if policy.MaxPending != 128 {
		t.Errorf("expected 128, got %d", policy.MaxPending)
	}
}

func TestBuildHedgePolicy_EmptyWindow_UsesDefault(t *testing.T) {
	cfg := HedgeConfig{Window: "", MaxPending: 64}
	policy, err := BuildHedgePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defaultPolicy := DefaultHedgeConfig()
	expected, _ := BuildHedgePolicy(defaultPolicy)
	if policy.Window != expected.Window {
		t.Errorf("expected default window %v, got %v", expected.Window, policy.Window)
	}
}

func TestBuildHedgePolicy_InvalidDuration(t *testing.T) {
	cfg := HedgeConfig{Window: "not-a-duration"}
	_, err := BuildHedgePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestBuildHedgePolicy_NegativeWindow_Error(t *testing.T) {
	cfg := HedgeConfig{Window: "-1s"}
	_, err := BuildHedgePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestBuildHedgePolicy_ZeroMaxPending_UsesDefault(t *testing.T) {
	cfg := HedgeConfig{Window: "2s", MaxPending: 0}
	policy, err := BuildHedgePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxPending <= 0 {
		t.Errorf("expected positive MaxPending from default fallback, got %d", policy.MaxPending)
	}
}
