package config

import (
	"fmt"
	"time"

	"github.com/yourorg/portwatch/internal/portscanner"
)

// FenceConfig holds user-facing configuration for the rate-limit fence.
type FenceConfig struct {
	// MaxEvents is the maximum events per Window before the fence trips.
	MaxEvents int `toml:"max_events" json:"max_events"`
	// Window is the rolling time window expressed as a duration string.
	Window string `toml:"window" json:"window"`
	// CooldownAfterFence is the duration the fence stays active once tripped.
	CooldownAfterFence string `toml:"cooldown_after_fence" json:"cooldown_after_fence"`
}

// DefaultFenceConfig returns a FenceConfig with sensible defaults.
func DefaultFenceConfig() FenceConfig {
	return FenceConfig{
		MaxEvents:          50,
		Window:             "1m",
		CooldownAfterFence: "2m",
	}
}

// BuildFencePolicy converts a FenceConfig into a portscanner.FencePolicy.
// It returns an error if any duration string is invalid or thresholds are
// non-positive.
func BuildFencePolicy(cfg FenceConfig) (portscanner.FencePolicy, error) {
	if cfg.MaxEvents <= 0 {
		return portscanner.FencePolicy{}, fmt.Errorf("fence: max_events must be > 0, got %d", cfg.MaxEvents)
	}

	window, err := time.ParseDuration(cfg.Window)
	if err != nil {
		return portscanner.FencePolicy{}, fmt.Errorf("fence: invalid window %q: %w", cfg.Window, err)
	}
	if window <= 0 {
		return portscanner.FencePolicy{}, fmt.Errorf("fence: window must be positive")
	}

	cooldown, err := time.ParseDuration(cfg.CooldownAfterFence)
	if err != nil {
		return portscanner.FencePolicy{}, fmt.Errorf("fence: invalid cooldown_after_fence %q: %w", cfg.CooldownAfterFence, err)
	}
	if cooldown <= 0 {
		return portscanner.FencePolicy{}, fmt.Errorf("fence: cooldown_after_fence must be positive")
	}

	return portscanner.FencePolicy{
		MaxEvents:          cfg.MaxEvents,
		Window:             window,
		CooldownAfterFence: cooldown,
	}, nil
}
