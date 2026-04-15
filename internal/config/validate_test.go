package config

import (
	"testing"
	"time"
)

func validConfig() *Config {
	c := DefaultConfig()
	c.NotifyOnOpen = true
	return c
}

func TestValidate_DefaultConfigIsValid(t *testing.T) {
	c := validConfig()
	if err := Validate(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_IntervalTooShort(t *testing.T) {
	c := validConfig()
	c.Interval = 100 * time.Millisecond
	err := Validate(c)
	if err == nil {
		t.Fatal("expected validation error for short interval")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestValidate_IntervalTooLong(t *testing.T) {
	c := validConfig()
	c.Interval = 25 * time.Hour
	if err := Validate(c); err == nil {
		t.Fatal("expected validation error for long interval")
	}
}

func TestValidate_EmptyWebhookURL(t *testing.T) {
	c := validConfig()
	c.WebhookURLs = []string{"https://example.com", ""}
	err := Validate(c)
	if err == nil {
		t.Fatal("expected validation error for empty webhook URL")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(ve.Errors))
	}
}

func TestValidate_DesktopNotifyWithoutEvents(t *testing.T) {
	c := validConfig()
	c.DesktopNotify = true
	c.NotifyOnOpen = false
	c.NotifyOnClose = false
	if err := Validate(c); err == nil {
		t.Fatal("expected validation error when desktop_notify set but no events enabled")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	c := validConfig()
	c.Interval = 10 * time.Millisecond
	c.WebhookURLs = []string{""}
	err := Validate(c)
	if err == nil {
		t.Fatal("expected multiple validation errors")
	}
	ve := err.(*ValidationError)
	if len(ve.Errors) < 2 {
		t.Fatalf("expected at least 2 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestIsValidationError(t *testing.T) {
	if IsValidationError(nil) {
		t.Fatal("nil should not be a ValidationError")
	}
	ve := &ValidationError{Errors: []string{"oops"}}
	if !IsValidationError(ve) {
		t.Fatal("expected true for *ValidationError")
	}
}
