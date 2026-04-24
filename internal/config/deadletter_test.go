package config

import (
	"testing"
)

func TestDefaultDeadLetterConfig_Valid(t *testing.T) {
	cfg := DefaultDeadLetterConfig()
	policy, err := BuildDeadLetterPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxSize != 256 {
		t.Errorf("expected MaxSize 256, got %d", policy.MaxSize)
	}
	if !policy.LogDropped {
		t.Error("expected LogDropped true")
	}
}

func TestBuildDeadLetterPolicy_ZeroMaxSize_UsesDefault(t *testing.T) {
	cfg := DeadLetterConfig{MaxSize: 0, LogDropped: false}
	policy, err := BuildDeadLetterPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxSize != 256 {
		t.Errorf("expected default MaxSize 256, got %d", policy.MaxSize)
	}
}

func TestBuildDeadLetterPolicy_CustomMaxSize(t *testing.T) {
	cfg := DeadLetterConfig{MaxSize: 512, LogDropped: true}
	policy, err := BuildDeadLetterPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxSize != 512 {
		t.Errorf("expected MaxSize 512, got %d", policy.MaxSize)
	}
}

func TestBuildDeadLetterPolicy_NegativeMaxSize_Error(t *testing.T) {
	cfg := DeadLetterConfig{MaxSize: -1}
	_, err := BuildDeadLetterPolicy(cfg)
	if err == nil {
		t.Fatal("expected error for negative MaxSize, got nil")
	}
}

func TestBuildDeadLetterPolicy_LogDropped_False(t *testing.T) {
	cfg := DeadLetterConfig{MaxSize: 64, LogDropped: false}
	policy, err := BuildDeadLetterPolicy(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.LogDropped {
		t.Error("expected LogDropped false")
	}
}
