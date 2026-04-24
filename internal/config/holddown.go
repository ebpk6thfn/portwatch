package config

import (
	"fmt"
	"time"

	"github.com/example/portwatch/internal/portscanner"
)

// HolddownConfig is the TOML-serialisable representation of hold-down
// timer settings.
type HolddownConfig struct {
	// Duration is a human-readable duration string, e.g. "10s", "1m".
	// An empty string or "0" disables hold-down.
	Duration string `toml:"duration"`
}

// DefaultHolddownConfig returns the recommended defaults.
func DefaultHolddownConfig() HolddownConfig {
	return HolddownConfig{
		Duration: "10s",
	}
}

// BuildHolddownPolicy converts a HolddownConfig into a
// portscanner.HolddownPolicy, validating the duration field.
func BuildHolddownPolicy(cfg HolddownConfig) (portscanner.HolddownPolicy, error) {
	if cfg.Duration == "" || cfg.Duration == "0" {
		return portscanner.HolddownPolicy{Duration: 0}, nil
	}

	d, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		return portscanner.HolddownPolicy{}, fmt.Errorf("holddown: invalid duration %q: %w", cfg.Duration, err)
	}
	if d < 0 {
		return portscanner.HolddownPolicy{}, fmt.Errorf("holddown: duration must be non-negative, got %s", cfg.Duration)
	}
	if d > 10*time.Minute {
		return portscanner.HolddownPolicy{}, fmt.Errorf("holddown: duration %s exceeds maximum of 10m", cfg.Duration)
	}

	return portscanner.HolddownPolicy{Duration: d}, nil
}
