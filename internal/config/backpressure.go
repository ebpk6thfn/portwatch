package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// BackpressureConfig holds user-facing configuration for backpressure limiting.
type BackpressureConfig struct {
	// HighWatermark is the queue depth that triggers backpressure (default 100).
	HighWatermark int `toml:"high_watermark"`
	// LowWatermark is the queue depth at which backpressure is released (default 25).
	LowWatermark int `toml:"low_watermark"`
	// CooldownPeriod is how long to stay in backpressure state before releasing (default 5s).
	CooldownPeriod string `toml:"cooldown_period"`
}

// DefaultBackpressureConfig returns safe defaults.
func DefaultBackpressureConfig() BackpressureConfig {
	return BackpressureConfig{
		HighWatermark:  100,
		LowWatermark:   25,
		CooldownPeriod: "5s",
	}
}

// BuildBackpressurePolicy converts a BackpressureConfig into a portscanner.BackpressurePolicy.
func BuildBackpressurePolicy(cfg BackpressureConfig) (portscanner.BackpressurePolicy, error) {
	if cfg.HighWatermark <= 0 {
		return portscanner.BackpressurePolicy{}, fmt.Errorf("high_watermark must be positive, got %d", cfg.HighWatermark)
	}
	if cfg.LowWatermark < 0 {
		return portscanner.BackpressurePolicy{}, fmt.Errorf("low_watermark must be non-negative, got %d", cfg.LowWatermark)
	}
	if cfg.LowWatermark >= cfg.HighWatermark {
		return portscanner.BackpressurePolicy{}, fmt.Errorf("low_watermark (%d) must be less than high_watermark (%d)", cfg.LowWatermark, cfg.HighWatermark)
	}

	cooldown := 5 * time.Second
	if cfg.CooldownPeriod != "" {
		d, err := time.ParseDuration(cfg.CooldownPeriod)
		if err != nil {
			return portscanner.BackpressurePolicy{}, fmt.Errorf("invalid cooldown_period %q: %w", cfg.CooldownPeriod, err)
		}
		if d < 0 {
			return portscanner.BackpressurePolicy{}, fmt.Errorf("cooldown_period must be non-negative")
		}
		cooldown = d
	}

	return portscanner.BackpressurePolicy{
		HighWatermark:  cfg.HighWatermark,
		LowWatermark:   cfg.LowWatermark,
		CooldownPeriod: cooldown,
	}, nil
}
