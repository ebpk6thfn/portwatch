package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// HedgeConfig holds the TOML/YAML-serialisable hedge configuration.
type HedgeConfig struct {
	// Window is a duration string, e.g. "3s".
	Window     string `toml:"window" yaml:"window"`
	MaxPending int    `toml:"max_pending" yaml:"max_pending"`
}

// DefaultHedgeConfig returns safe defaults.
func DefaultHedgeConfig() HedgeConfig {
	p := portscanner.DefaultHedgePolicy()
	return HedgeConfig{
		Window:     p.Window.String(),
		MaxPending: p.MaxPending,
	}
}

// BuildHedgePolicy converts a HedgeConfig into a portscanner.HedgePolicy.
func BuildHedgePolicy(cfg HedgeConfig) (portscanner.HedgePolicy, error) {
	defaults := portscanner.DefaultHedgePolicy()

	window := defaults.Window
	if cfg.Window != "" {
		d, err := time.ParseDuration(cfg.Window)
		if err != nil {
			return portscanner.HedgePolicy{}, fmt.Errorf("hedge: invalid window %q: %w", cfg.Window, err)
		}
		if d < 0 {
			return portscanner.HedgePolicy{}, fmt.Errorf("hedge: window must not be negative")
		}
		window = d
	}

	maxPending := cfg.MaxPending
	if maxPending <= 0 {
		maxPending = defaults.MaxPending
	}

	return portscanner.HedgePolicy{
		Window:     window,
		MaxPending: maxPending,
	}, nil
}
