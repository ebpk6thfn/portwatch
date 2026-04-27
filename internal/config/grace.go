package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// GraceConfig holds the user-facing configuration for the startup grace period.
type GraceConfig struct {
	// WindowDuration is a parseable duration string, e.g. "5s" or "10s".
	// Defaults to "5s" when empty.
	WindowDuration string `toml:"window" yaml:"window"`
}

// DefaultGraceConfig returns a GraceConfig with a 5-second startup window.
func DefaultGraceConfig() GraceConfig {
	return GraceConfig{
		WindowDuration: "5s",
	}
}

// BuildGracePolicy converts a GraceConfig into a portscanner.GracePolicy.
// An empty WindowDuration falls back to the default (5s).
func BuildGracePolicy(cfg GraceConfig) (portscanner.GracePolicy, error) {
	raw := cfg.WindowDuration
	if raw == "" {
		return portscanner.DefaultGracePolicy(), nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return portscanner.GracePolicy{}, fmt.Errorf("grace: invalid window %q: %w", raw, err)
	}
	if d < 0 {
		return portscanner.GracePolicy{}, fmt.Errorf("grace: window must be non-negative, got %v", d)
	}

	return portscanner.GracePolicy{Window: d}, nil
}
