package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// DrainConfig holds user-facing drain configuration.
type DrainConfig struct {
	MaxBuffer int    `toml:"max_buffer"`
	MaxAge    string `toml:"max_age"`
}

// DefaultDrainConfig returns safe defaults.
func DefaultDrainConfig() DrainConfig {
	return DrainConfig{
		MaxBuffer: 64,
		MaxAge:    "10s",
	}
}

// BuildDrainPolicy converts DrainConfig into a portscanner.DrainPolicy.
func BuildDrainPolicy(c DrainConfig) (portscanner.DrainPolicy, error) {
	p := portscanner.DefaultDrainPolicy()
	if c.MaxBuffer > 0 {
		p.MaxBuffer = c.MaxBuffer
	}
	if c.MaxAge != "" {
		d, err := time.ParseDuration(c.MaxAge)
		if err != nil {
			return p, fmt.Errorf("drain: invalid max_age %q: %w", c.MaxAge, err)
		}
		if d <= 0 {
			return p, fmt.Errorf("drain: max_age must be positive")
		}
		p.MaxAge = d
	}
	return p, nil
}
