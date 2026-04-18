package config

import (
	"errors"
	"time"
)

// SummaryConfig holds configuration for periodic summary reporting.
type SummaryConfig struct {
	Enabled  bool   `toml:"enabled"`
	Interval string `toml:"interval"`
}

// DefaultSummaryConfig returns a sensible default summary configuration.
func DefaultSummaryConfig() SummaryConfig {
	return SummaryConfig{
		Enabled:  false,
		Interval: "5m",
	}
}

// SummaryPolicy is the resolved form of SummaryConfig.
type SummaryPolicy struct {
	Enabled  bool
	Interval time.Duration
}

// BuildSummaryPolicy parses and validates a SummaryConfig.
func BuildSummaryPolicy(c SummaryConfig) (SummaryPolicy, error) {
	if !c.Enabled {
		return SummaryPolicy{Enabled: false}, nil
	}
	if c.Interval == "" {
		return SummaryPolicy{}, errors.New("summary interval must not be empty when enabled")
	}
	d, err := time.ParseDuration(c.Interval)
	if err != nil {
		return SummaryPolicy{}, fmt.Errorf("invalid summary interval %q: %w", c.Interval, err)
	}
	if d < 30*time.Second {
		return SummaryPolicy{}, errors.New("summary interval must be at least 30s")
	}
	return SummaryPolicy{Enabled: true, Interval: d}, nil
}
