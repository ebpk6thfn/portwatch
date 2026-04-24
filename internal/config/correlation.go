package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// CorrelationConfig holds user-facing correlation settings.
type CorrelationConfig struct {
	WindowSeconds int `toml:"window_seconds"`
	MinCount      int `toml:"min_count"`
}

// DefaultCorrelationConfig returns safe defaults.
func DefaultCorrelationConfig() CorrelationConfig {
	return CorrelationConfig{
		WindowSeconds: 30,
		MinCount:      3,
	}
}

// BuildCorrelationPolicy converts config to a portscanner policy.
func BuildCorrelationPolicy(cfg CorrelationConfig) (portscanner.CorrelationPolicy, error) {
	if cfg.WindowSeconds <= 0 {
		return portscanner.CorrelationPolicy{}, errors.New("correlation window_seconds must be positive")
	}
	if cfg.MinCount <= 0 {
		return portscanner.CorrelationPolicy{}, fmt.Errorf("correlation min_count must be >= 1, got %d", cfg.MinCount)
	}
	return portscanner.CorrelationPolicy{
		Window:   time.Duration(cfg.WindowSeconds) * time.Second,
		MinCount: cfg.MinCount,
	}, nil
}
