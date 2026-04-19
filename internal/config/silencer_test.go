package config

import (
	"testing"
)

func TestDefaultSilencerConfig_Empty(t *testing.T) {
	cfg := DefaultSilencerConfig()
	if len(cfg.Rules) != 0 {
		t.Fatalf("expected no rules, got %d", len(cfg.Rules))
	}
}

func TestBuildSilenceRules_Valid(t *testing.T) {
	cfg := SilencerConfig{
		Rules: []SilenceRule{
			{Port: 8080, Duration: "30m"},
			{Port: 443, Duration: "2h"},
		},
	}
	rules, err := BuildSilenceRules(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Port != 8080 {
		t.Errorf("expected port 8080")
	}
}

func TestBuildSilenceRules_InvalidPort(t *testing.T) {
	cfg := SilencerConfig{
		Rules: []SilenceRule{{Port: 0, Duration: "10m"}},
	}
	_, err := BuildSilenceRules(cfg)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestBuildSilenceRules_EmptyDuration(t *testing.T) {
	cfg := SilencerConfig{
		Rules: []SilenceRule{{Port: 80, Duration: ""}},
	}
	_, err := BuildSilenceRules(cfg)
	if err == nil {
		t.Fatal("expected error for empty duration")
	}
}

func TestBuildSilenceRules_InvalidDurationString(t *testing.T) {
	cfg := SilencerConfig{
		Rules: []SilenceRule{{Port: 80, Duration: "notaduration"}},
	}
	_, err := BuildSilenceRules(cfg)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestBuildSilenceRules_NegativeDuration(t *testing.T) {
	cfg := SilencerConfig{
		Rules: []SilenceRule{{Port: 80, Duration: "-5m"}},
	}
	_, err := BuildSilenceRules(cfg)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestBuildSilenceRules_Empty(t *testing.T) {
	cfg := DefaultSilencerConfig()
	rules, err := BuildSilenceRules(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules")
	}
}
