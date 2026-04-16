package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func makeBaselineEntry(proto, ip string, port uint16) Entry {
	return Entry{Protocol: proto, IP: ip, Port: port}
}

func TestNewBaseline_ContainsEntries(t *testing.T) {
	entries := []Entry{
		makeBaselineEntry("tcp", "0.0.0.0", 80),
		makeBaselineEntry("tcp", "0.0.0.0", 443),
	}
	b := NewBaseline(entries)
	for _, e := range entries {
		if !b.Contains(e) {
			t.Errorf("expected baseline to contain %s", e.Key())
		}
	}
}

func TestNewBaseline_DoesNotContainOther(t *testing.T) {
	b := NewBaseline([]Entry{makeBaselineEntry("tcp", "0.0.0.0", 80)})
	other := makeBaselineEntry("tcp", "0.0.0.0", 8080)
	if b.Contains(other) {
		t.Error("baseline should not contain unlisted entry")
	}
}

func TestSaveAndLoadBaseline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	original := NewBaseline([]Entry{
		makeBaselineEntry("tcp", "127.0.0.1", 22),
		makeBaselineEntry("udp", "0.0.0.0", 53),
	})

	if err := SaveBaseline(path, original); err != nil {
		t.Fatalf("SaveBaseline: %v", err)
	}

	loaded, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}

	for key := range original.Entries {
		if !loaded.Entries[key] {
			t.Errorf("loaded baseline missing key %s", key)
		}
	}
}

func TestLoadBaseline_MissingFile(t *testing.T) {
	b, err := LoadBaseline("/nonexistent/path/baseline.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if b == nil || b.Entries == nil {
		t.Error("expected empty baseline, got nil")
	}
}

func TestLoadBaseline_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json{"), 0644)
	_, err := LoadBaseline(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
