package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// ReaperConfig holds raw configuration values for the Reaper.
type ReaperConfig struct {
	MaxAge   string `toml:"max_age"   json:"max_age"`
	Interval string `toml:"interval"  json:"interval"`
}

// DefaultReaperConfig returns safe defaults.
func DefaultReaperConfig() ReaperConfig {
	return ReaperConfig{
		MaxAge:   "10m",
		Interval: "2m",
	}
}

// BuildReaperPolicy parses a ReaperConfig into a portscanner.ReaperPolicy.
func BuildReaperPolicy(cfg ReaperConfig) (portscanner.ReaperPolicy, error) {
	policy := portscanner.DefaultReaperPolicy()

	if cfg.MaxAge != "" {
		d, err := time.ParseDuration(cfg.MaxAge)
		if err != nil {
			return policy, fmt.Errorf("reaper: invalid max_age %q: %w", cfg.MaxAge, err)
		}
		if d <= 0 {
			return policy, fmt.Errorf("reaper: max_age must be positive")
		}
		policy.MaxAge = d
	}

	if cfg.Interval != "" {
		d, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			return policy, fmt.Errorf("reaper: invalid interval %q: %w", cfg.Interval, err)
		}
		if d <= 0 {
			return policy, fmt.Errorf("reaper: interval must be positive")
		}
		policy.Interval = d
	}

	return policy, nil
}
