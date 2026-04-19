package config

import (
	"errors"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// QuotaConfig holds user-facing quota configuration.
type QuotaConfig struct {
	WindowStr string `toml:"window"`
	MaxHigh   int    `toml:"max_high"`
	MaxMedium int    `toml:"max_medium"`
	MaxLow    int    `toml:"max_low"`
}

// DefaultQuotaConfig returns sensible defaults.
func DefaultQuotaConfig() QuotaConfig {
	return QuotaConfig{
		WindowStr: "1h",
		MaxHigh:   50,
		MaxMedium: 100,
		MaxLow:    200,
	}
}

// BuildQuotaPolicy converts QuotaConfig into a portscanner.QuotaPolicy.
func BuildQuotaPolicy(cfg QuotaConfig) (portscanner.QuotaPolicy, error) {
	if cfg.WindowStr == "" {
		cfg = DefaultQuotaConfig()
	}
	w, err := time.ParseDuration(cfg.WindowStr)
	if err != nil {
		return portscanner.QuotaPolicy{}, errors.New("quota: invalid window: " + err.Error())
	}
	if w <= 0 {
		return portscanner.QuotaPolicy{}, errors.New("quota: window must be positive")
	}
	if cfg.MaxHigh <= 0 || cfg.MaxMedium <= 0 || cfg.MaxLow <= 0 {
		return portscanner.QuotaPolicy{}, errors.New("quota: max values must be positive")
	}
	return portscanner.QuotaPolicy{
		Window:    w,
		MaxHigh:   cfg.MaxHigh,
		MaxMedium: cfg.MaxMedium,
		MaxLow:    cfg.MaxLow,
	}, nil
}
