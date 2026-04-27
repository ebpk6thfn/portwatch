package config

import (
	"testing"
	"time"
)

func TestDefaultGraceConfig_Valid(t *testing.T) {
	cfg := DefaultGraceConfig()
	policy, err := BuildGracePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != 5*time.Second {
		t.Fatalf("expected 5s, got %v", policy.Window)
	}
}

func TestBuildGracePolicy_EmptyWindow_UsesDefault(t *testing.T) {
	cfg := GraceConfig{WindowDuration: ""}
	policy, err := BuildGracePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window <= 0 {
		t.Fatalf("expected positive default window, got %v", policy.Window)
	}
}

func TestBuildGracePolicy_CustomValues(t *testing.T) {
	cfg := GraceConfig{WindowDuration: "30s"}
	policy, err := BuildGracePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != 30*time.Second {
		t.Fatalf("expected 30s, got %v", policy.Window)
	}
}

func TestBuildGracePolicy_ZeroWindow_Allowed(t *testing.T) {
	cfg := GraceConfig{WindowDuration: "0s"}
	policy, err := BuildGracePolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error for zero window: %v", err)
	}
	if policy.Window != 0 {
		t.Fatalf("expected 0 window, got %v", policy.Window)
	}
}

func TestBuildGracePolicy_InvalidDuration(t *testing.T) {
	cfg := GraceConfig{WindowDuration: "not-a-duration"}
	_, err := BuildGracePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestBuildGracePolicy_NegativeDuration_Error(t *testing.T) {
	cfg := GraceConfig{WindowDuration: "-5s"}
	_, err := BuildGracePolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
}
