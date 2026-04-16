package portscanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempCheckpointPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestCheckpoint_LoadMissingFile(t *testing.T) {
	cs := NewCheckpointStore(tempCheckpointPath(t))
	cp, err := cs.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if !cp.LastScan.IsZero() {
		t.Errorf("expected zero LastScan, got %v", cp.LastScan)
	}
	if cp.ScanCount != 0 || cp.EventCount != 0 {
		t.Errorf("expected zero counters, got scan=%d event=%d", cp.ScanCount, cp.EventCount)
	}
}

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	cs := NewCheckpointStore(tempCheckpointPath(t))
	now := time.Now().UTC().Truncate(time.Second)
	cp := Checkpoint{LastScan: now, ScanCount: 5, EventCount: 42}
	if err := cs.Save(cp); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !loaded.LastScan.Equal(now) {
		t.Errorf("LastScan mismatch: want %v got %v", now, loaded.LastScan)
	}
	if loaded.ScanCount != 5 || loaded.EventCount != 42 {
		t.Errorf("counter mismatch: scan=%d event=%d", loaded.ScanCount, loaded.EventCount)
	}
}

func TestCheckpoint_Record_IncrementsCounters(t *testing.T) {
	cs := NewCheckpointStore(tempCheckpointPath(t))
	if err := cs.Record(3); err != nil {
		t.Fatalf("first Record: %v", err)
	}
	if err := cs.Record(7); err != nil {
		t.Fatalf("second Record: %v", err)
	}
	cp, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cp.ScanCount != 2 {
		t.Errorf("ScanCount: want 2 got %d", cp.ScanCount)
	}
	if cp.EventCount != 10 {
		t.Errorf("EventCount: want 10 got %d", cp.EventCount)
	}
	if cp.LastScan.IsZero() {
		t.Error("LastScan should not be zero after Record")
	}
}

func TestCheckpoint_InvalidJSON(t *testing.T) {
	path := tempCheckpointPath(t)
	if err := os.WriteFile(path, []byte("not-json{"), 0o600); err != nil {
		t.Fatal(err)
	}
	cs := NewCheckpointStore(path)
	_, err := cs.Load()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
