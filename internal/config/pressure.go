package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// PressureConfig holds user-facing configuration for the pressure gauge.
type PressureConfig struct {
	LowWatermark  int    `toml:"low_watermark"`
	HighWatermark int    `toml:"high_watermark"`
	Window        string `toml:"window"`
}

// DefaultPressureConfig returns a PressureConfig with sensible defaults.
func DefaultPressureConfig() PressureConfig {
	return PressureConfig{
		LowWatermark:  10,
		HighWatermark: 50,
		Window:        "30s",
	}
}

// BuildPressurePolicy converts a PressureConfig into a portscanner.PressurePolicy.
func BuildPressurePolicy(cfg PressureConfig) (portscanner.PressurePolicy, error) {
	if cfg.LowWatermark < 0 {
		return portscanner.PressurePolicy{}, fmt.Errorf("pressure: low_watermark must be >= 0")
	}
	if cfg.HighWatermark <= 0 {
		return portscanner.PressurePolicy{}, fmt.Errorf("pressure: high_watermark must be > 0")
	}
	if cfg.LowWatermark >= cfg.HighWatermark {
		return portscanner.PressurePolicy{}, fmt.Errorf("pressure: low_watermark must be less than high_watermark")
	}

	window := 30 * time.Second
	if cfg.Window != "" {
		d, err := time.ParseDuration(cfg.Window)
		if err != nil {
			return portscanner.PressurePolicy{}, fmt.Errorf("pressure: invalid window %q: %w", cfg.Window, err)
		}
		if d <= 0 {
			return portscanner.PressurePolicy{}, fmt.Errorf("pressure: window must be positive")
		}
		window = d
	}

	return portscanner.PressurePolicy{
		LowWatermark:  cfg.LowWatermark,
		HighWatermark: cfg.HighWatermark,
		Window:        window,
	}, nil
}
