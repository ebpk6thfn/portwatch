package config

import (
	"testing"
	"time"
)

func TestDefaultReaperConfig_Valid(t *testing.T) {
	cfg := DefaultReaperConfig()
	policy, err := BuildReaperPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxAge <= 0 {
		t.Error("expected positive MaxAge")
	}
	if policy.Interval <= 0 {
		t.Error("expected positive Interval")
	}
}

func TestBuildReaperPolicy_CustomValues(t *testing.T) {
	cfg := ReaperConfig{MaxAge: "20m", Interval: "5m"}
	policy, err := BuildReaperPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxAge != 20*time.Minute {
		t.Errorf("expected MaxAge=20m, got %v", policy.MaxAge)
	}
	if policy.Interval != 5*time.Minute {
		t.Errorf("expected Interval=5m, got %v", policy.Interval)
	}
}

func TestBuildReaperPolicy_InvalidMaxAge(t *testing.T) {
	cfg := ReaperConfig{MaxAge: "not-a-duration", Interval: "1m"}
	_, err := BuildReaperPolicy(cfg)
	if err == nil {
		t.Error("expected error for invalid max_age")
	}
}

func TestBuildReaperPolicy_InvalidInterval(t *testing.T) {
	cfg := ReaperConfig{MaxAge: "10m", Interval: "bad"}
	_, err := BuildReaperPolicy(cfg)
	if err == nil {
		t.Error("expected error for invalid interval")
	}
}

func TestBuildReaperPolicy_NegativeMaxAge_Error(t *testing.T) {
	cfg := ReaperConfig{MaxAge: "-5m", Interval: "1m"}
	_, err := BuildReaperPolicy(cfg)
	if err == nil {
		t.Error("expected error for negative max_age")
	}
}

func TestBuildReaperPolicy_EmptyStrings_UsesDefaults(t *testing.T) {
	cfg := ReaperConfig{}
	policy, err := BuildReaperPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxAge <= 0 {
		t.Error("expected positive MaxAge from defaults")
	}
}
