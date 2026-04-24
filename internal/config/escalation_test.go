package config

import (
	"testing"
	"time"
)

func TestDefaultEscalationConfig_Valid(t *testing.T) {
	cfg := DefaultEscalationConfig()
	policy, err := BuildEscalationPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.CountThreshold != 3 {
		t.Errorf("expected CountThreshold=3, got %d", policy.CountThreshold)
	}
	if policy.Window != 2*time.Minute {
		t.Errorf("expected Window=2m, got %v", policy.Window)
	}
}

func TestBuildEscalationPolicy_CustomValues(t *testing.T) {
	cfg := EscalationConfig{CountThreshold: 5, Window: "30s"}
	policy, err := BuildEscalationPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.CountThreshold != 5 {
		t.Errorf("expected 5, got %d", policy.CountThreshold)
	}
	if policy.Window != 30*time.Second {
		t.Errorf("expected 30s, got %v", policy.Window)
	}
}

func TestBuildEscalationPolicy_ZeroThreshold_Error(t *testing.T) {
	cfg := EscalationConfig{CountThreshold: 0, Window: "1m"}
	_, err := BuildEscalationPolicy(cfg)
	if err == nil {
		t.Error("expected error for zero count_threshold")
	}
}

func TestBuildEscalationPolicy_InvalidDuration(t *testing.T) {
	cfg := EscalationConfig{CountThreshold: 2, Window: "notaduration"}
	_, err := BuildEscalationPolicy(cfg)
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestBuildEscalationPolicy_EmptyWindow_UsesDefault(t *testing.T) {
	cfg := EscalationConfig{CountThreshold: 2, Window: ""}
	policy, err := BuildEscalationPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != 2*time.Minute {
		t.Errorf("expected default window 2m, got %v", policy.Window)
	}
}

func TestBuildEscalationPolicy_NegativeWindow_Error(t *testing.T) {
	cfg := EscalationConfig{CountThreshold: 2, Window: "-1m"}
	_, err := BuildEscalationPolicy(cfg)
	if err == nil {
		t.Error("expected error for negative window")
	}
}
