package config

import (
	"testing"
)

func TestDefaultLedgerConfig_Valid(t *testing.T) {
	cfg := DefaultLedgerConfig()
	if cfg.MaxSize <= 0 {
		t.Errorf("expected positive default MaxSize, got %d", cfg.MaxSize)
	}
	if !cfg.Enabled {
		t.Error("expected default ledger to be enabled")
	}
}

func TestBuildLedgerPolicy_DefaultValues(t *testing.T) {
	cfg := DefaultLedgerConfig()
	pol, err := BuildLedgerPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.MaxSize != cfg.MaxSize {
		t.Errorf("expected MaxSize %d, got %d", cfg.MaxSize, pol.MaxSize)
	}
	if pol.Enabled != cfg.Enabled {
		t.Errorf("expected Enabled %v, got %v", cfg.Enabled, pol.Enabled)
	}
}

func TestBuildLedgerPolicy_ZeroMaxSize_Unlimited(t *testing.T) {
	cfg := LedgerConfig{MaxSize: 0, Enabled: true}
	pol, err := BuildLedgerPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.MaxSize != 0 {
		t.Errorf("expected MaxSize 0 (unlimited), got %d", pol.MaxSize)
	}
}

func TestBuildLedgerPolicy_NegativeMaxSize_Error(t *testing.T) {
	cfg := LedgerConfig{MaxSize: -1, Enabled: true}
	_, err := BuildLedgerPolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative MaxSize")
	}
}

func TestBuildLedgerPolicy_Disabled(t *testing.T) {
	cfg := LedgerConfig{MaxSize: 100, Enabled: false}
	pol, err := BuildLedgerPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Enabled {
		t.Error("expected policy to be disabled")
	}
}
