package config

import (
	"errors"

	"github.com/user/portwatch/internal/portscanner"
)

// ShadowConfig holds TOML-serialisable shadow-mode settings.
type ShadowConfig struct {
	Enabled    bool `toml:"enabled"`
	LogDropped bool `toml:"log_dropped"`
	MaxDropped int  `toml:"max_dropped"`
}

// DefaultShadowConfig returns sensible defaults (shadow mode off).
func DefaultShadowConfig() ShadowConfig {
	return ShadowConfig{
		Enabled:    false,
		LogDropped: true,
		MaxDropped: 512,
	}
}

// BuildShadowPolicy validates cfg and converts it to a portscanner.ShadowPolicy.
func BuildShadowPolicy(cfg ShadowConfig) (portscanner.ShadowPolicy, error) {
	if cfg.MaxDropped < 0 {
		return portscanner.ShadowPolicy{}, errors.New("shadow: max_dropped must be >= 0")
	}
	return portscanner.ShadowPolicy{
		Enabled:    cfg.Enabled,
		LogDropped: cfg.LogDropped,
		MaxDropped: cfg.MaxDropped,
	}, nil
}
