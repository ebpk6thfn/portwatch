package config

import (
	"errors"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// DispatchConfig holds user-facing configuration for the event dispatcher.
type DispatchConfig struct {
	Workers    int    `toml:"workers"`
	QueueDepth int    `toml:"queue_depth"`
	TimeoutStr string `toml:"timeout"`
}

// DefaultDispatchConfig returns a DispatchConfig with sensible defaults.
func DefaultDispatchConfig() DispatchConfig {
	return DispatchConfig{
		Workers:    4,
		QueueDepth: 256,
		TimeoutStr: "5s",
	}
}

// BuildDispatchPolicy converts a DispatchConfig into a portscanner.DispatchPolicy.
func BuildDispatchPolicy(cfg DispatchConfig) (portscanner.DispatchPolicy, error) {
	if cfg.Workers <= 0 {
		return portscanner.DispatchPolicy{}, errors.New("dispatch: workers must be positive")
	}
	if cfg.QueueDepth <= 0 {
		return portscanner.DispatchPolicy{}, errors.New("dispatch: queue_depth must be positive")
	}
	timeoutStr := cfg.TimeoutStr
	if timeoutStr == "" {
		timeoutStr = DefaultDispatchConfig().TimeoutStr
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return portscanner.DispatchPolicy{}, errors.New("dispatch: invalid timeout: " + err.Error())
	}
	if timeout <= 0 {
		return portscanner.DispatchPolicy{}, errors.New("dispatch: timeout must be positive")
	}
	return portscanner.DispatchPolicy{
		Workers:    cfg.Workers,
		QueueDepth: cfg.QueueDepth,
		Timeout:    timeout,
	}, nil
}
