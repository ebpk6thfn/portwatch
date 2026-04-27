package config

import (
	"fmt"

	"github.com/user/portwatch/internal/portscanner"
)

// JournalConfig holds user-facing configuration for the event journal.
type JournalConfig struct {
	Enabled    bool `toml:"enabled" yaml:"enabled"`
	MaxEntries int  `toml:"max_entries" yaml:"max_entries"`
}

// DefaultJournalConfig returns a safe default journal configuration.
func DefaultJournalConfig() JournalConfig {
	return JournalConfig{
		Enabled:    true,
		MaxEntries: 500,
	}
}

// BuildJournalPolicy converts a JournalConfig into a portscanner.JournalPolicy.
// It returns an error if any value is out of range.
func BuildJournalPolicy(cfg JournalConfig) (portscanner.JournalPolicy, error) {
	if cfg.MaxEntries < 0 {
		return portscanner.JournalPolicy{}, fmt.Errorf("journal: max_entries must be non-negative, got %d", cfg.MaxEntries)
	}
	max := cfg.MaxEntries
	if max == 0 {
		max = DefaultJournalConfig().MaxEntries
	}
	return portscanner.JournalPolicy{
		MaxEntries: max,
	}, nil
}
