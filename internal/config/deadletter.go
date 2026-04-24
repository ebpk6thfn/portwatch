package config

import (
	"errors"
	"fmt"
)

// DeadLetterConfig holds configuration for the dead-letter queue.
type DeadLetterConfig struct {
	// MaxSize is the maximum number of dead-letter entries to retain.
	// Defaults to 256 when zero.
	MaxSize int `toml:"max_size" yaml:"max_size"`

	// LogDropped controls whether dropped events are logged to stderr.
	LogDropped bool `toml:"log_dropped" yaml:"log_dropped"`
}

// DefaultDeadLetterConfig returns a sensible default configuration.
func DefaultDeadLetterConfig() DeadLetterConfig {
	return DeadLetterConfig{
		MaxSize:    256,
		LogDropped: true,
	}
}

// DeadLetterPolicy is the resolved, validated policy used at runtime.
type DeadLetterPolicy struct {
	MaxSize    int
	LogDropped bool
}

// BuildDeadLetterPolicy validates and converts a DeadLetterConfig into a
// DeadLetterPolicy ready for use by the runtime.
func BuildDeadLetterPolicy(cfg DeadLetterConfig) (DeadLetterPolicy, error) {
	if cfg.MaxSize < 0 {
		return DeadLetterPolicy{}, fmt.Errorf("%w: dead_letter.max_size must not be negative",
			errors.New("validation error"))
	}
	size := cfg.MaxSize
	if size == 0 {
		size = 256
	}
	return DeadLetterPolicy{
		MaxSize:    size,
		LogDropped: cfg.LogDropped,
	}, nil
}
