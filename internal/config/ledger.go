package config

import (
	"errors"
	"fmt"
)

// LedgerConfig holds configuration for the event ledger.
type LedgerConfig struct {
	// MaxSize is the maximum number of unique port keys tracked.
	// 0 means unlimited.
	MaxSize int `toml:"max_size" json:"max_size"`

	// Enabled controls whether the ledger is active.
	Enabled bool `toml:"enabled" json:"enabled"`
}

// DefaultLedgerConfig returns a sensible default ledger configuration.
func DefaultLedgerConfig() LedgerConfig {
	return LedgerConfig{
		MaxSize: 4096,
		Enabled: true,
	}
}

// LedgerPolicy is the resolved policy used at runtime.
type LedgerPolicy struct {
	MaxSize int
	Enabled bool
}

// BuildLedgerPolicy validates and converts a LedgerConfig into a LedgerPolicy.
func BuildLedgerPolicy(cfg LedgerConfig) (LedgerPolicy, error) {
	if cfg.MaxSize < 0 {
		return LedgerPolicy{}, fmt.Errorf("%w: ledger max_size must be >= 0, got %d",
			ErrInvalidConfig, cfg.MaxSize)
	}
	return LedgerPolicy{
		MaxSize: cfg.MaxSize,
		Enabled: cfg.Enabled,
	}, nil
}

// ErrInvalidConfig is a sentinel used by config validation helpers.
var ErrInvalidConfig = errors.New("invalid config")
