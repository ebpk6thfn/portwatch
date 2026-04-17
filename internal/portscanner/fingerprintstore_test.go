package portscanner

import (
	"os"
	"path/filepath"
	"testing"
)

func tempFPPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "fingerprint.json")
}

func TestFingerprintStore_LoadMissingFile(t *testing.T) {
	store := NewFingerprintStore(tempFPPath(t))
	fp, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fp.Hash != "" {
		t.Errorf("expected empty hash, got %s", fp.Hash)
	}
}

func TestFingerprintStore_SaveAndLoad(t *testing.T) {
	path := tempFPPath(t)
	store := NewFingerprintStore(path)
	fb := NewFingerprintBuilder()
	original := fb.Build([]Entry{makeFPEntry("tcp", 80), makeFPEntry("udp", 53)})
	if err := store.Save(original); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Hash != original.Hash {
		t.Errorf("hash mismatch: got %s want %s", loaded.Hash, original.Hash)
	}
	if len(loaded.PortSet) != len(original.PortSet) {
		t.Errorf("port set length mismatch: got %d want %d", len(loaded.PortSet), len(original.PortSet))
	}
}

func TestFingerprintStore_OverwriteExisting(t *testing.T) {
	path := tempFPPath(t)
	store := NewFingerprintStore(path)
	fb := NewFingerprintBuilder()
	first := fb.Build([]Entry{makeFPEntry("tcp", 80)})
	second := fb.Build([]Entry{makeFPEntry("tcp", 443)})
	_ = store.Save(first)
	_ = store.Save(second)
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Hash != second.Hash {
		t.Errorf("expected second hash %s, got %s", second.Hash, loaded.Hash)
	}
}

func TestFingerprintStore_InvalidJSON(t *testing.T) {
	path := tempFPPath(t)
	_ = os.WriteFile(path, []byte("not-json"), 0o600)
	store := NewFingerprintStore(path)
	_, err := store.Load()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
