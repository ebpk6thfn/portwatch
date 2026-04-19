package config

import (
	"testing"
	"time"
)

func TestDefaultQuotaConfig_Valid(t *testing.T) {
	cfg := DefaultQuotaConfig()
	policy, err := BuildQuotaPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window != time.Hour {
		t.Errorf("expected 1h window, got %v", policy.Window)
	}
}

func TestBuildQuotaPolicy_EmptyWindow_UsesDefault(t *testing.T) {
	cfg := QuotaConfig{} // zero value
	policy, err := BuildQuotaPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Window <= 0 {
		t.Error("expected positive window")
	}
}

func TestBuildQuotaPolicy_InvalidWindow(t *testing.T) {
	cfg := QuotaConfig{WindowStr: "not-a-duration", MaxHigh: 10, MaxMedium: 10, MaxLow: 10}
	_, err := BuildQuotaPolicy(cfg)
	if err == nil {
		t.Fatal("expected error for invalid window")
	}
}

func TestBuildQuotaPolicy_NegativeWindow(t *testing.T) {
	cfg := QuotaConfig{WindowStr: "-1h", MaxHigh: 10, MaxMedium: 10, MaxLow: 10}
	_, err := BuildQuotaPolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestBuildQuotaPolicy_ZeroMax(t *testing.T) {
	tests := []struct {
		name string
		cfg  QuotaConfig
	}{
		{"zero MaxHigh", QuotaConfig{WindowStr: "1h", MaxHigh: 0, MaxMedium: 10, MaxLow: 10}},
		{"zero MaxMedium", QuotaConfig{WindowStr: "1h", MaxHigh: 10, MaxMedium: 0, MaxLow: 10}},
		{"zero MaxLow", QuotaConfig{WindowStr: "1h", MaxHigh: 10, MaxMedium: 10, MaxLow: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildQuotaPolicy(tt.cfg)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

func TestBuildQuotaPolicy_CustomValues(t *testing.T) {
	cfg := QuotaConfig{WindowStr: "30m", MaxHigh: 5, MaxMedium: 20, MaxLow: 100}
	policy, err := BuildQuotaPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxHigh != 5 {
		t.Errorf("expected MaxHigh 5, got %d", policy.MaxHigh)
	}
	if policy.Window != 30*time.Minute {
		t.Errorf("expected 30m window, got %v", policy.Window)
	}
}
