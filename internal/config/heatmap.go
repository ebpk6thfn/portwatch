package config

import (
	"errors"
	"time"
)

// HeatmapConfig holds configuration for the port activity heatmap.
type HeatmapConfig struct {
	// Window is the duration string over which hits are counted, e.g. "5m".
	Window string `toml:"window" yaml:"window"`
	// TopN controls how many entries are returned by the heatmap Top query.
	TopN int `toml:"top_n" yaml:"top_n"`
}

// HeatmapPolicy is the resolved, validated form of HeatmapConfig.
type HeatmapPolicy struct {
	Window time.Duration
	TopN   int
}

// DefaultHeatmapConfig returns sensible defaults for the heatmap.
func DefaultHeatmapConfig() HeatmapConfig {
	return HeatmapConfig{
		Window: "5m",
		TopN:   10,
	}
}

// BuildHeatmapPolicy validates and converts a HeatmapConfig into a HeatmapPolicy.
func BuildHeatmapPolicy(cfg HeatmapConfig) (HeatmapPolicy, error) {
	if cfg.Window == "" {
		cfg.Window = DefaultHeatmapConfig().Window
	}

	w, err := time.ParseDuration(cfg.Window)
	if err != nil {
		return HeatmapPolicy{}, errors.New("heatmap: invalid window duration: " + err.Error())
	}
	if w <= 0 {
		return HeatmapPolicy{}, errors.New("heatmap: window must be positive")
	}

	topN := cfg.TopN
	if topN <= 0 {
		topN = DefaultHeatmapConfig().TopN
	}

	return HeatmapPolicy{
		Window: w,
		TopN:   topN,
	}, nil
}
