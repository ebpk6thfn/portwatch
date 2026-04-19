package config

import "time"

// PausableConfig controls automatic pause/resume behaviour for the scanner.
type PausableConfig struct {
	// AutoPauseOnBurst pauses event forwarding when a burst is detected.
	AutoPauseOnBurst bool `toml:"auto_pause_on_burst"`
	// AutoResumeDuration is how long the scanner stays paused before auto-resuming.
	// Zero means it never auto-resumes.
	AutoResumeDuration time.Duration `toml:"auto_resume_duration"`
}

// DefaultPausableConfig returns safe defaults.
func DefaultPausableConfig() PausableConfig {
	return PausableConfig{
		AutoPauseOnBurst:   false,
		AutoResumeDuration: 30 * time.Second,
	}
}

// BuildPausablePolicy validates and returns the config, falling back to
// defaults for zero values.
func BuildPausablePolicy(cfg PausableConfig) (PausableConfig, error) {
	if cfg.AutoResumeDuration < 0 {
		return cfg, &ValidationError{Field: "auto_resume_duration", Msg: "must not be negative"}
	}
	if cfg.AutoResumeDuration == 0 {
		cfg.AutoResumeDuration = DefaultPausableConfig().AutoResumeDuration
	}
	return cfg, nil
}
