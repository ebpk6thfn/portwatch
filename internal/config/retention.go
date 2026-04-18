package config

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// RetentionConfig holds configuration for the event retention store.
type RetentionConfig struct {
	MaxAgeDuration string `toml:"max_age" json:"max_age"`
	MaxCount       int    `toml:"max_count" json:"max_count"`
}

// DefaultRetentionConfig returns sensible defaults.
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		MaxAgeDuration: "1h",
		MaxCount:       1000,
	}
}

// BuildRetentionPolicy parses the config and returns a RetentionPolicy.
func BuildRetentionPolicy(rc RetentionConfig) (portscanner.RetentionPolicy, error) {
	policy := portscanner.RetentionPolicy{
		MaxCount: rc.MaxCount,
	}
	if rc.MaxAgeDuration != "" {
		d, err := time.ParseDuration(rc.MaxAgeDuration)
		if err != nil {
			return policy, fmt.Errorf("retention: invalid max_age %q: %w", rc.MaxAgeDuration, err)
		}
		policy.MaxAge = d
	}
	return policy, nil
}
