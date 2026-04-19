package config

import (
	"fmt"
	"time"
)

// CircuitBreakerConfig holds configuration for the circuit breaker notifier wrapper.
type CircuitBreakerConfig struct {
	Enabled      bool   `toml:"enabled"`
	Threshold    int    `toml:"threshold"`
	RecoveryWait string `toml:"recovery_wait"`
}

// DefaultCircuitBreakerConfig returns safe defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:      true,
		Threshold:    5,
		RecoveryWait: "30s",
	}
}

// CircuitBreakerPolicy is the resolved, typed form of CircuitBreakerConfig.
type CircuitBreakerPolicy struct {
	Enabled      bool
	Threshold    int
	RecoveryWait time.Duration
}

// BuildCircuitBreakerPolicy parses and validates a CircuitBreakerConfig.
func BuildCircuitBreakerPolicy(cfg CircuitBreakerConfig) (CircuitBreakerPolicy, error) {
	if cfg.Threshold <= 0 {
		return CircuitBreakerPolicy{}, fmt.Errorf("circuit breaker threshold must be > 0, got %d", cfg.Threshold)
	}
	d, err := time.ParseDuration(cfg.RecoveryWait)
	if err != nil {
		return CircuitBreakerPolicy{}, fmt.Errorf("circuit breaker recovery_wait: %w", err)
	}
	if d <= 0 {
		return CircuitBreakerPolicy{}, fmt.Errorf("circuit breaker recovery_wait must be positive")
	}
	return CircuitBreakerPolicy{
		Enabled:      cfg.Enabled,
		Threshold:    cfg.Threshold,
		RecoveryWait: d,
	}, nil
}
