package config

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds all runtime configuration for portwatch.
type Config struct {
	// Interval is how often the daemon scans for port changes.
	Interval time.Duration `toml:"interval"`
	// WebhookURL is the HTTP endpoint to POST change events to.
	// Leave empty to disable webhook notifications.
	WebhookURL string `toml:"webhook_url"`
	// DesktopNotify enables native desktop notifications.
	DesktopNotify bool `toml:"desktop_notify"`
	// ExcludePorts is a list of local ports to ignore during scanning.
	ExcludePorts []uint16 `toml:"exclude_ports"`
	// ExcludeLoopback skips loopback-bound ports when true.
	ExcludeLoopback bool `toml:"exclude_loopback"`
	// Protocols restricts scanning to the listed protocols.
	// Accepted values: "tcp", "udp". Empty means both.
	Protocols []string `toml:"protocols"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Interval:        5 * time.Second,
		DesktopNotify:   false,
		ExcludeLoopback: false,
		Protocols:       []string{"tcp", "udp"},
	}
}

// Load reads a TOML config file from path. If path is empty the
// default configuration is returned without error.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if err := Validate(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
