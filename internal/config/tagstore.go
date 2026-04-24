package config

import (
	"fmt"
	"time"
)

// TagRule defines a static tag assignment for a specific port/protocol key.
type TagRule struct {
	// Key is the port key in the form "proto:port", e.g. "tcp:80".
	Key string `toml:"key"`
	// Tags is the list of string labels to attach.
	Tags []string `toml:"tags"`
	// TTL is an optional duration string. Empty means tags never expire.
	TTL string `toml:"ttl"`
}

// TagStoreConfig holds the top-level configuration for the tag store.
type TagStoreConfig struct {
	Rules []TagRule `toml:"rules"`
}

// DefaultTagStoreConfig returns an empty configuration (no static rules).
func DefaultTagStoreConfig() TagStoreConfig {
	return TagStoreConfig{}
}

// ParsedTagRule is a TagRule with the TTL already parsed into a duration.
type ParsedTagRule struct {
	Key  string
	Tags []string
	TTL  time.Duration // zero means no expiry
}

// BuildTagRules validates and parses a TagStoreConfig into a slice of
// ParsedTagRule values ready for use with portscanner.TagStore.
func BuildTagRules(cfg TagStoreConfig) ([]ParsedTagRule, error) {
	out := make([]ParsedTagRule, 0, len(cfg.Rules))
	for i, r := range cfg.Rules {
		if r.Key == "" {
			return nil, fmt.Errorf("tag rule %d: key must not be empty", i)
		}
		if len(r.Tags) == 0 {
			return nil, fmt.Errorf("tag rule %d (%s): tags must not be empty", i, r.Key)
		}
		var ttl time.Duration
		if r.TTL != "" {
			var err error
			ttl, err = time.ParseDuration(r.TTL)
			if err != nil {
				return nil, fmt.Errorf("tag rule %d (%s): invalid ttl %q: %w", i, r.Key, r.TTL, err)
			}
			if ttl <= 0 {
				return nil, fmt.Errorf("tag rule %d (%s): ttl must be positive", i, r.Key)
			}
		}
		out = append(out, ParsedTagRule{Key: r.Key, Tags: r.Tags, TTL: ttl})
	}
	return out, nil
}
