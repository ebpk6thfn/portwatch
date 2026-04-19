package config

import (
	"fmt"
	"time"
)

// SilenceRule configures a port silence window.
type SilenceRule struct {
	Port     uint16 `toml:"port"`
	Duration string `toml:"duration"`
}

// SilencerConfig holds all silence rules.
type SilencerConfig struct {
	Rules []SilenceRule `toml:"rules"`
}

// DefaultSilencerConfig returns an empty silencer config.
func DefaultSilencerConfig() SilencerConfig {
	return SilencerConfig{}
}

// ParsedSilenceRule is a validated rule with a parsed duration.
type ParsedSilenceRule struct {
	Port     uint16
	Duration time.Duration
}

// BuildSilenceRules parses and validates SilencerConfig rules.
func BuildSilenceRules(cfg SilencerConfig) ([]ParsedSilenceRule, error) {
	var out []ParsedSilenceRule
	for _, r := range cfg.Rules {
		if r.Port == 0 {
			return nil, fmt.Errorf("silence rule has invalid port 0")
		}
		if r.Duration == "" {
			return nil, fmt.Errorf("silence rule for port %d has empty duration", r.Port)
		}
		d, err := time.ParseDuration(r.Duration)
		if err != nil {
			return nil, fmt.Errorf("silence rule for port %d: invalid duration %q: %w", r.Port, r.Duration, err)
		}
		if d <= 0 {
			return nil, fmt.Errorf("silence rule for port %d: duration must be positive", r.Port)
		}
		out = append(out, ParsedSilenceRule{Port: r.Port, Duration: d})
	}
	return out, nil
}
