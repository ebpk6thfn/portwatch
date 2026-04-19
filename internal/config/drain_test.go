package config

import (
	"testing"
	"time"
)

func TestDefaultDrainConfig_Valid(t *testing.T) {
	c := DefaultDrainConfig()
	p, err := BuildDrainPolicy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxBuffer != 64 {
		t.Fatalf("expected MaxBuffer 64, got %d", p.MaxBuffer)
	}
	if p.MaxAge != 10*time.Second {
		t.Fatalf("expected MaxAge 10s, got %v", p.MaxAge)
	}
}

func TestBuildDrainPolicy_CustomValues(t *testing.T) {
	c := DrainConfig{MaxBuffer: 32, MaxAge: "30s"}
	p, err := BuildDrainPolicy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxBuffer != 32 {
		t.Fatalf("expected 32, got %d", p.MaxBuffer)
	}
	if p.MaxAge != 30*time.Second {
		t.Fatalf("expected 30s, got %v", p.MaxAge)
	}
}

func TestBuildDrainPolicy_InvalidDuration(t *testing.T) {
	c := DrainConfig{MaxBuffer: 10, MaxAge: "not-a-duration"}
	_, err := BuildDrainPolicy(c)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestBuildDrainPolicy_ZeroMaxBuffer_UsesDefault(t *testing.T) {
	c := DrainConfig{MaxBuffer: 0, MaxAge: "5s"}
	p, err := BuildDrainPolicy(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxBuffer != 64 {
		t.Fatalf("expected default MaxBuffer 64, got %d", p.MaxBuffer)
	}
}

func TestBuildDrainPolicy_NegativeAge_Error(t *testing.T) {
	c := DrainConfig{MaxBuffer: 10, MaxAge: "-1s"}
	_, err := BuildDrainPolicy(c)
	if err == nil {
		t.Fatal("expected error for negative max_age")
	}
}
