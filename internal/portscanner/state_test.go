package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStatePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestStateStore_LoadMissingFile(t *testing.T) {
	store := NewStateStore(tempStatePath(t))
	got, err := store.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(got))
	}
}

func TestStateStore_SaveAndLoad(t *testing.T) {
	store := NewStateStore(tempStatePath(t))

	snapshot := map[string]Entry{
		"tcp:8080": {Protocol: "tcp", Port: 8080, LocalAddr: "0.0.0.0"},
		"tcp:443":  {Protocol: "tcp", Port: 443, LocalAddr: "0.0.0.0"},
	}

	if err := store.Save(snapshot); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded) != len(snapshot) {
		t.Fatalf("expected %d entries, got %d", len(snapshot), len(loaded))
	}
	for k := range snapshot {
		if _, ok := loaded[k]; !ok {
			t.Errorf("missing key %q after round-trip", k)
		}
	}
}

func TestStateStore_OverwriteExisting(t *testing.T) {
	path := tempStatePath(t)
	store := NewStateStore(path)

	first := map[string]Entry{
		"tcp:9000": {Protocol: "tcp", Port: 9000, LocalAddr: "127.0.0.1"},
	}
	if err := store.Save(first); err != nil {
		t.Fatalf("first Save() error: %v", err)
	}

	second := map[string]Entry{
		"udp:5353": {Protocol: "udp", Port: 5353, LocalAddr: "0.0.0.0"},
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("second Save() error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 entry after overwrite, got %d", len(loaded))
	}
	if _, ok := loaded["udp:5353"]; !ok {
		t.Error("expected udp:5353 in loaded state")
	}
}

func TestStateStore_InvalidJSON(t *testing.T) {
	path := tempStatePath(t)
	if err := os.WriteFile(path, []byte("not-json{"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewStateStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
