package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ScanInterval != 5*time.Second {
		t.Errorf("expected 5s interval, got %v", cfg.ScanInterval)
	}
	if cfg.Desktop {
		t.Error("expected desktop notifications off by default")
	}
	if len(cfg.ProcNetPaths) == 0 {
		t.Error("expected non-empty default proc_net_paths")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeTempConfig(t, `
scan_interval: 10s
webhook_url: "https://example.com/hook"
desktop_notifications: true
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ScanInterval != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.ScanInterval)
	}
	if cfg.WebhookURL != "https://example.com/hook" {
		t.Errorf("unexpected webhook url: %s", cfg.WebhookURL)
	}
	if !cfg.Desktop {
		t.Error("expected desktop notifications enabled")
	}
}

func TestLoad_InvalidInterval(t *testing.T) {
	path := writeTempConfig(t, "scan_interval: -1s\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for negative scan_interval")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/portwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
