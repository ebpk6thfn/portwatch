package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// StaggerConfig holds user-facing configuration for the stagger feature.
type StaggerConfig struct {
	WindowStr string `toml:"window" json:"window"`
	MaxSlot   int    `toml:"max_slot" json:"max_slot"`
}

// DefaultStaggerConfig returns a default stagger configuration.
func DefaultStaggerConfig() StaggerConfig {
	return StaggerConfig{
		WindowStr: "5s",
		MaxSlot:   10,
	}
}

// BuildStaggerPolicy converts a StaggerConfig into a portscanner.StaggerPolicy.
func BuildStaggerPolicy(cfg StaggerConfig) (portscanner.StaggerPolicy, error) {
	defaults := portscanner.DefaultStaggerPolicy()

	window := defaults.Window
	if cfg.WindowStr != "" {
		d, err := time.ParseDuration(cfg.WindowStr)
		if err != nil {
			return portscanner.StaggerPolicy{}, fmt.Errorf("stagger: invalid window %q: %w", cfg.WindowStr, err)
		}
		if d < 0 {
			return portscanner.StaggerPolicy{}, fmt.Errorf("stagger: window must be non-negative, got %v", d)
		}
		window = d
	}

	maxSlot := cfg.MaxSlot
	if maxSlot <= 0 {
		maxSlot = defaults.MaxSlot
	}

	return portscanner.StaggerPolicy{
		Window:  window,
		MaxSlot: maxSlot,
	}, nil
}
