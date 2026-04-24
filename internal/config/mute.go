package config

import (
	"fmt"
	"time"

	"github.com/yourorg/portwatch/internal/portscanner"
)

// MuteConfig holds user-facing configuration for the mute feature.
type MuteConfig struct {
	// Duration is how long a mute lasts, e.g. "30m", "1h".
	Duration string `toml:"duration" yaml:"duration"`
}

// DefaultMuteConfig returns conservative defaults.
func DefaultMuteConfig() MuteConfig {
	return MuteConfig{
		Duration: "30m",
	}
}

// BuildMutePolicy converts a MuteConfig into a portscanner.MutePolicy.
// Returns an error if the duration string is invalid or non-positive.
func BuildMutePolicy(cfg MuteConfig) (portscanner.MutePolicy, error) {
	if cfg.Duration == "" {
		def := portscanner.DefaultMutePolicy()
		return def, nil
	}

	d, err := time.ParseDuration(cfg.Duration)
	if err != nil {
		return portscanner.MutePolicy{}, fmt.Errorf("mute: invalid duration %q: %w", cfg.Duration, err)
	}
	if d <= 0 {
		return portscanner.MutePolicy{}, fmt.Errorf("mute: duration must be positive, got %s", cfg.Duration)
	}

	return portscanner.MutePolicy{Duration: d}, nil
}
