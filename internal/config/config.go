package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full portwatch daemon configuration.
type Config struct {
	ScanInterval time.Duration `yaml:"scan_interval"`
	WebhookURL   string        `yaml:"webhook_url"`
	Desktop      bool          `yaml:"desktop_notifications"`
	ProcNetPaths []string      `yaml:"proc_net_paths"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ScanInterval: 5 * time.Second,
		Desktop:      false,
		ProcNetPaths: []string{
			"/proc/net/tcp",
			"/proc/net/tcp6",
		},
	}
}

// Load reads a YAML config file from path and merges it over the defaults.
// If path is empty the defaults are returned as-is.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if cfg.ScanInterval <= 0 {
		return nil, fmt.Errorf("config: scan_interval must be positive")
	}

	return cfg, nil
}
