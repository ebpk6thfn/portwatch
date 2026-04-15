package config

import (
	"errors"
	"fmt"
	"time"
)

// ValidationError holds a list of validation issues found in a Config.
type ValidationError struct {
	Errors []string
}

func (v *ValidationError) Error() string {
	if len(v.Errors) == 1 {
		return fmt.Sprintf("config validation error: %s", v.Errors[0])
	}
	return fmt.Sprintf("config validation errors: %v", v.Errors)
}

// Validate checks that c contains sensible values and returns a
// *ValidationError describing every problem found, or nil if c is valid.
func Validate(c *Config) error {
	var errs []string

	if c.Interval < 500*time.Millisecond {
		errs = append(errs, "interval must be at least 500ms")
	}
	if c.Interval > 24*time.Hour {
		errs = append(errs, "interval must not exceed 24h")
	}

	for i, u := range c.WebhookURLs {
		if u == "" {
			errs = append(errs, fmt.Sprintf("webhook_urls[%d] must not be empty", i))
		}
	}

	if c.DesktopNotify && c.NotifyOnOpen == false && c.NotifyOnClose == false {
		errs = append(errs, "desktop_notify is enabled but neither notify_on_open nor notify_on_close is set")
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// IsValidationError returns true when err is (or wraps) a *ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
