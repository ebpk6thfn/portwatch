package config

import (
	"testing"
)

func TestDefaultTagStoreConfig_Empty(t *testing.T) {
	cfg := DefaultTagStoreConfig()
	if len(cfg.Rules) != 0 {
		t.Fatalf("expected no rules, got %d", len(cfg.Rules))
	}
}

func TestBuildTagRules_Valid(t *testing.T) {
	cfg := TagStoreConfig{
		Rules: []TagRule{
			{Key: "tcp:80", Tags: []string{"http"}, TTL: ""},
			{Key: "tcp:443", Tags: []string{"tls", "web"}, TTL: "1h"},
		},
	}
	rules, err := BuildTagRules(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[1].TTL.Hours() != 1 {
		t.Fatalf("expected 1h TTL, got %v", rules[1].TTL)
	}
}

func TestBuildTagRules_EmptyKey_Error(t *testing.T) {
	cfg := TagStoreConfig{
		Rules: []TagRule{
			{Key: "", Tags: []string{"x"}},
		},
	}
	_, err := BuildTagRules(cfg)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestBuildTagRules_EmptyTags_Error(t *testing.T) {
	cfg := TagStoreConfig{
		Rules: []TagRule{
			{Key: "tcp:22", Tags: []string{}},
		},
	}
	_, err := BuildTagRules(cfg)
	if err == nil {
		t.Fatal("expected error for empty tags slice")
	}
}

func TestBuildTagRules_InvalidTTL_Error(t *testing.T) {
	cfg := TagStoreConfig{
		Rules: []TagRule{
			{Key: "udp:53", Tags: []string{"dns"}, TTL: "not-a-duration"},
		},
	}
	_, err := BuildTagRules(cfg)
	if err == nil {
		t.Fatal("expected error for invalid TTL string")
	}
}

func TestBuildTagRules_NegativeTTL_Error(t *testing.T) {
	cfg := TagStoreConfig{
		Rules: []TagRule{
			{Key: "tcp:8080", Tags: []string{"proxy"}, TTL: "-5m"},
		},
	}
	_, err := BuildTagRules(cfg)
	if err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestBuildTagRules_ZeroTTL_NoExpiry(t *testing.T) {
	cfg := TagStoreConfig{
		Rules: []TagRule{
			{Key: "tcp:22", Tags: []string{"ssh"}, TTL: ""},
		},
	}
	rules, err := BuildTagRules(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rules[0].TTL != 0 {
		t.Fatalf("expected zero TTL for empty string, got %v", rules[0].TTL)
	}
}
