package config

import (
	"testing"
	"time"
)

func TestDefaultPausableConfig_Valid(t *testing.T) {
	cfg := DefaultPausableConfig()
	if cfg.AutoPauseOnBurst {
		t.Error("expected auto_pause_on_burst false by default")
	}
	if cfg.AutoResumeDuration != 30*time.Second {
		t.Errorf("expected 30s default, got %v", cfg.AutoResumeDuration)
	}
}

func TestBuildPausablePolicy_NegativeDuration(t *testing.T) {
	cfg := PausableConfig{AutoResumeDuration: -1 * time.Second}
	_, err := BuildPausablePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestBuildPausablePolicy_ZeroDurationFallsToDefault(t *testing.T) {
	cfg := PausableConfig{AutoResumeDuration: 0}
	out, err := BuildPausablePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AutoResumeDuration != 30*time.Second {
		t.Errorf("expected fallback to 30s, got %v", out.AutoResumeDuration)
	}
}

func TestBuildPausablePolicy_CustomDuration(t *testing.T) {
	cfg := PausableConfig{AutoResumeDuration: 2 * time.Minute}
	out, err := BuildPausablePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AutoResumeDuration != 2*time.Minute {
		t.Errorf("expected 2m, got %v", out.AutoResumeDuration)
	}
}
