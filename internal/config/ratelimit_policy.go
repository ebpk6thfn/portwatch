package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// RateLimitPolicyConfig holds serialisable rate limit policy settings.
type RateLimitPolicyConfig struct {
	DefaultCooldown  string            `toml:"default_cooldown"`
	HighCooldown     string            `toml:"high_cooldown"`
	MediumCooldown   string            `toml:"medium_cooldown"`
	LowCooldown      string            `toml:"low_cooldown"`
	ProtocolOverride map[string]string `toml:"protocol_override"`
}

// DefaultRateLimitPolicyConfig returns safe defaults.
func DefaultRateLimitPolicyConfig() RateLimitPolicyConfig {
	return RateLimitPolicyConfig{
		DefaultCooldown: "30s",
		HighCooldown:    "5s",
		MediumCooldown:  "15s",
		LowCooldown:     "60s",
		ProtocolOverride: map[string]string{},
	}
}

// BuildRateLimitPolicy converts config into a portscanner.RateLimitPolicy.
func BuildRateLimitPolicy(c RateLimitPolicyConfig) (portscanner.RateLimitPolicy, error) {
	parse := func(s string) (time.Duration, error) {
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return d, nil
	}
	def, err := parse(c.DefaultCooldown)
	if err != nil { return portscanner.RateLimitPolicy{}, err }
	hi, err := parse(c.HighCooldown)
	if err != nil { return portscanner.RateLimitPolicy{}, err }
	med, err := parse(c.MediumCooldown)
	if err != nil { return portscanner.RateLimitPolicy{}, err }
	lo, err := parse(c.LowCooldown)
	if err != nil { return portscanner.RateLimitPolicy{}, err }

	overrides := make(map[string]time.Duration, len(c.ProtocolOverride))
	for proto, ds := range c.ProtocolOverride {
		d, err := parse(ds)
		if err != nil { return portscanner.RateLimitPolicy{}, err }
		overrides[proto] = d
	}
	return portscanner.RateLimitPolicy{
		DefaultCooldown:  def,
		HighCooldown:     hi,
		MediumCooldown:   med,
		LowCooldown:      lo,
		ProtocolOverride: overrides,
	}, nil
}
