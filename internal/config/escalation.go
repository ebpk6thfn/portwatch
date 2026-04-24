package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// EscalationConfig holds user-facing escalation configuration.
type EscalationConfig struct {
	// CountThreshold is the number of events within Window that triggers
	// severity escalation. Defaults to 3.
	CountThreshold int    `toml:"count_threshold"`
	Window         string `toml:"window"`
}

// DefaultEscalationConfig returns defaults suitable for most deployments.
func DefaultEscalationConfig() EscalationConfig {
	return EscalationConfig{
		CountThreshold: 3,
		Window:         "2m",
	}
}

// BuildEscalationPolicy converts EscalationConfig into a portscanner policy.
func BuildEscalationPolicy(cfg EscalationConfig) (portscanner.EscalationPolicy, error) {
	if cfg.CountThreshold <= 0 {
		return portscanner.EscalationPolicy{}, errors.New("escalation count_threshold must be > 0")
	}

	win := cfg.Window
	if win == "" {
		win = DefaultEscalationConfig().Window
	}

	d, err := time.ParseDuration(win)
	if err != nil {
		return portscanner.EscalationPolicy{}, fmt.Errorf("invalid escalation window %q: %w", win, err)
	}
	if d <= 0 {
		return portscanner.EscalationPolicy{}, errors.New("escalation window must be positive")
	}

	return portscanner.EscalationPolicy{
		CountThreshold: cfg.CountThreshold,
		Window:         d,
	}, nil
}
