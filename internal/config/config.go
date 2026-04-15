package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds all runtime configuration for portwatch.
type Config struct {
	// Interval between consecutive port scans.
	Interval time.Duration `toml:"interval"`

	// WebhookURLs is a list of HTTP endpoints to POST change events to.
	WebhookURLs []string `toml:"webhook_urls"`

	// DesktopNotify enables desktop (OS) notifications.
	DesktopNotify bool `toml:"desktop_notify"`

	// NotifyOnOpen triggers a notification when a new port is detected.
	NotifyOnOpen bool `toml:"notify_on_open"`

	// NotifyOnClose triggers a notification when a port disappears.
	NotifyOnClose bool `toml:"notify_on_close"`
}

// DefaultConfig returns a Config populated with safe defaults.
func DefaultConfig() *Config {
	return &Config{
		Interval:      5 * time.Second,
		WebhookURLs:   []string{},
		DesktopNotify: false,
		NotifyOnOpen:  true,
		NotifyOnClose: true,
	}
}

// Load reads a TOML config file from path and merges it over DefaultConfig.
// If path is empty the defaults are returned unchanged.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse toml: %w", err)
	}

	return cfg, nil
}
