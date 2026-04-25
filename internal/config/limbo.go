package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// LimboConfig holds user-facing configuration for the limbo feature.
type LimboConfig struct {
	// Window is how long an event waits for confirmation before being discarded.
	// Accepts a Go duration string, e.g. "5s", "500ms".
	Window  string `toml:"window" yaml:"window" json:"window"`
	MaxSize int    `toml:"max_size" yaml:"max_size" json:"max_size"`
}

// DefaultLimboConfig returns a sensible default configuration.
func DefaultLimboConfig() LimboConfig {
	return LimboConfig{
		Window:  "5s",
		MaxSize: 256,
	}
}

// BuildLimboPolicy converts a LimboConfig into a portscanner.LimboPolicy.
func BuildLimboPolicy(cfg LimboConfig) (portscanner.LimboPolicy, error) {
	if cfg.Window == "" {
		cfg = DefaultLimboConfig()
	}

	d, err := time.ParseDuration(cfg.Window)
	if err != nil {
		return portscanner.LimboPolicy{}, fmt.Errorf("limbo: invalid window %q: %w", cfg.Window, err)
	}
	if d <= 0 {
		return portscanner.LimboPolicy{}, errors.New("limbo: window must be positive")
	}

	maxSize := cfg.MaxSize
	if maxSize <= 0 {
		maxSize = DefaultLimboConfig().MaxSize
	}

	return portscanner.LimboPolicy{
		Window:  d,
		MaxSize: maxSize,
	}, nil
}
