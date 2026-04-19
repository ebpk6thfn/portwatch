package config

import (
	"errors"
	"time"
)

// ProbeConfig holds configuration for the port liveness prober.
type ProbeConfig struct {
	Enabled        bool   `toml:"enabled"`
	TimeoutSeconds int    `toml:"timeout_seconds"`
	OnlyOnOpen     bool   `toml:"only_on_open"`
}

// DefaultProbeConfig returns a safe default ProbeConfig.
func DefaultProbeConfig() ProbeConfig {
	return ProbeConfig{
		Enabled:        false,
		TimeoutSeconds: 2,
		OnlyOnOpen:     true,
	}
}

// ProbePolicy is the resolved, typed form of ProbeConfig.
type ProbePolicy struct {
	Enabled    bool
	Timeout    time.Duration
	OnlyOnOpen bool
}

// BuildProbePolicy validates and converts a ProbeConfig into a ProbePolicy.
func BuildProbePolicy(c ProbeConfig) (ProbePolicy, error) {
	if c.TimeoutSeconds <= 0 {
		return ProbePolicy{}, errors.New("probe timeout_seconds must be positive")
	}
	if c.TimeoutSeconds > 30 {
		return ProbePolicy{}, errors.New("probe timeout_seconds must not exceed 30")
	}
	return ProbePolicy{
		Enabled:    c.Enabled,
		Timeout:    time.Duration(c.TimeoutSeconds) * time.Second,
		OnlyOnOpen: c.OnlyOnOpen,
	}, nil
}
