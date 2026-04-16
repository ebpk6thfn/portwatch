package portscanner

import (
	"encoding/json"
	"os"
	"time"
)

// Checkpoint records the last successful scan time and event count
// so the daemon can resume gracefully after restart.
type Checkpoint struct {
	LastScan   time.Time `json:"last_scan"`
	EventCount int64     `json:"event_count"`
	ScanCount  int64     `json:"scan_count"`
}

// CheckpointStore persists and loads Checkpoint data to/from disk.
type CheckpointStore struct {
	path string
}

// NewCheckpointStore creates a CheckpointStore that reads/writes to path.
func NewCheckpointStore(path string) *CheckpointStore {
	return &CheckpointStore{path: path}
}

// Save writes the checkpoint to disk atomically.
func (cs *CheckpointStore) Save(cp Checkpoint) error {
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	tmp := cs.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, cs.path)
}

// Load reads the checkpoint from disk. Returns a zero Checkpoint if the
// file does not exist.
func (cs *CheckpointStore) Load() (Checkpoint, error) {
	data, err := os.ReadFile(cs.path)
	if os.IsNotExist(err) {
		return Checkpoint{}, nil
	}
	if err != nil {
		return Checkpoint{}, err
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return Checkpoint{}, err
	}
	return cp, nil
}

// Record updates the checkpoint with a new scan timestamp and increments counters.
func (cs *CheckpointStore) Record(eventsEmitted int64) error {
	cp, err := cs.Load()
	if err != nil {
		return err
	}
	cp.LastScan = time.Now().UTC()
	cp.ScanCount++
	cp.EventCount += eventsEmitted
	return cs.Save(cp)
}
