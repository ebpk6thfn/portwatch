package config

import (
	"testing"
	"time"
)

func TestDefaultRetentionConfig_Valid(t *testing.T) {
	rc := DefaultRetentionConfig()
	policy, err := BuildRetentionPolicy(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxAge != time.Hour {
		t.Errorf("expected 1h, got %v", policy.MaxAge)
	}
	if policy.MaxCount != 1000 {
		t.Errorf("expected 1000, got %d", policy.MaxCount)
	}
}

func TestBuildRetentionPolicy_CustomValues(t *testing.T) {
	rc := RetentionConfig{MaxAgeDuration: "30m", MaxCount: 250}
	policy, err := BuildRetentionPolicy(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxAge != 30*time.Minute {
		t.Errorf("expected 30m, got %v", policy.MaxAge)
	}
	if policy.MaxCount != 250 {
		t.Errorf("expected 250, got %d", policy.MaxCount)
	}
}

func TestBuildRetentionPolicy_InvalidDuration(t *testing.T) {
	rc := RetentionConfig{MaxAgeDuration: "notaduration", MaxCount: 100}
	_, err := BuildRetentionPolicy(rc)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestBuildRetentionPolicy_EmptyAge_NoMaxAge(t *testing.T) {
	rc := RetentionConfig{MaxAgeDuration: "", MaxCount: 50}
	policy, err := BuildRetentionPolicy(rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.MaxAge != 0 {
		t.Errorf("expected zero MaxAge, got %v", policy.MaxAge)
	}
}
