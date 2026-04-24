package config

import (
	"testing"
	"time"
)

func TestDefaultCorrelationConfig_Valid(t *testing.T) {
	cfg := DefaultCorrelationConfig()
	policy, err := BuildCorrelationPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != 30*time.Second {
		t.Errorf("expected 30s window, got %s", policy.Window)
	}
	if policy.MinCount != 3 {
		t.Errorf("expected MinCount=3, got %d", policy.MinCount)
	}
}

func TestBuildCorrelationPolicy_CustomValues(t *testing.T) {
	cfg := CorrelationConfig{WindowSeconds: 60, MinCount: 5}
	policy, err := BuildCorrelationPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != 60*time.Second {
		t.Errorf("expected 60s, got %s", policy.Window)
	}
	if policy.MinCount != 5 {
		t.Errorf("expected MinCount=5, got %d", policy.MinCount)
	}
}

func TestBuildCorrelationPolicy_ZeroWindow_Error(t *testing.T) {
	cfg := CorrelationConfig{WindowSeconds: 0, MinCount: 3}
	_, err := BuildCorrelationPolicy(cfg)
	if err == nil {
		t.Error("expected error for zero window")
	}
}

func TestBuildCorrelationPolicy_NegativeWindow_Error(t *testing.T) {
	cfg := CorrelationConfig{WindowSeconds: -5, MinCount: 3}
	_, err := BuildCorrelationPolicy(cfg)
	if err == nil {
		t.Error("expected error for negative window")
	}
}

func TestBuildCorrelationPolicy_ZeroMinCount_Error(t *testing.T) {
	cfg := CorrelationConfig{WindowSeconds: 10, MinCount: 0}
	_, err := BuildCorrelationPolicy(cfg)
	if err == nil {
		t.Error("expected error for zero min_count")
	}
}

func TestBuildCorrelationPolicy_NegativeMinCount_Error(t *testing.T) {
	cfg := CorrelationConfig{WindowSeconds: 10, MinCount: -1}
	_, err := BuildCorrelationPolicy(cfg)
	if err == nil {
		t.Error("expected error for negative min_count")
	}
}
