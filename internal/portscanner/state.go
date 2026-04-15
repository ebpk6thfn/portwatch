package portscanner

import (
	"encoding/json"
	"os"
	"sync"
)

// StateStore persists the last known port snapshot to disk so that
// portwatch can detect changes across restarts.
type StateStore struct {
	mu   sync.RWMutex
	path string
}

// NewStateStore creates a StateStore backed by the given file path.
func NewStateStore(path string) *StateStore {
	return &StateStore{path: path}
}

// Load reads the persisted snapshot from disk. Returns an empty map if
// the file does not yet exist.
func (s *StateStore) Load() (map[string]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return make(map[string]Entry), nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	m := make(map[string]Entry, len(entries))
	for _, e := range entries {
		m[e.Key()] = e
	}
	return m, nil
}

// Save writes the provided snapshot map to disk atomically.
func (s *StateStore) Save(snapshot map[string]Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := make([]Entry, 0, len(snapshot))
	for _, e := range snapshot {
		entries = append(entries, e)
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
