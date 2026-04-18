package portscanner

import (
	"testing"
	"time"
)

func TestRateLimitPolicy_DefaultCooldown(t *testing.T) {
	p := DefaultRateLimitPolicy()
	if p.DefaultCooldown != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", p.DefaultCooldown)
	}
}

func TestRateLimitPolicy_CooldownFor_High(t *testing.T) {
	p := DefaultRateLimitPolicy()
	d := p.CooldownFor("high", "tcp")
	if d != 5*time.Second {
		t.Fatalf("expected 5s for high, got %v", d)
	}
}

func TestRateLimitPolicy_CooldownFor_Medium(t *testing.T) {
	p := DefaultRateLimitPolicy()
	d := p.CooldownFor("medium", "tcp")
	if d != 15*time.Second {
		t.Fatalf("expected 15s for medium, got %v", d)
	}
}

func TestRateLimitPolicy_CooldownFor_Low(t *testing.T) {
	p := DefaultRateLimitPolicy()
	d := p.CooldownFor("low", "udp")
	if d != 60*time.Second {
		t.Fatalf("expected 60s for low, got %v", d)
	}
}

func TestRateLimitPolicy_CooldownFor_ProtocolOverride(t *testing.T) {
	p := DefaultRateLimitPolicy()
	p.ProtocolOverride["udp"] = 2 * time.Second
	d := p.CooldownFor("high", "udp")
	if d != 2*time.Second {
		t.Fatalf("expected 2s for udp override, got %v", d)
	}
}

func TestRateLimitPolicy_CooldownFor_Unknown(t *testing.T) {
	p := DefaultRateLimitPolicy()
	d := p.CooldownFor("unknown", "tcp")
	if d != p.DefaultCooldown {
		t.Fatalf("expected default cooldown for unknown severity, got %v", d)
	}
}
