package portscanner

import (
	"encoding/json"
	"errors"
	"os"
)

// FingerprintRecord is a persisted fingerprint with metadata.
type FingerprintRecord struct {
	Hash    string   `json:"hash"`
	PortSet []string `json:"port_set"`
}

// FingerprintStore persists and loads fingerprints to/from disk.
type FingerprintStore struct {
	path string
}

// NewFingerprintStore returns a FingerprintStore backed by the given path.
func NewFingerprintStore(path string) *FingerprintStore {
	return &FingerprintStore{path: path}
}

// Save writes the fingerprint to disk.
func (fs *FingerprintStore) Save(fp Fingerprint) error {
	rec := FingerprintRecord{Hash: fp.Hash, PortSet: fp.PortSet}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.path, data, 0o600)
}

// Load reads the fingerprint from disk. Returns zero Fingerprint if file missing.
func (fs *FingerprintStore) Load() (Fingerprint, error) {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Fingerprint{}, nil
		}
		return Fingerprint{}, err
	}
	var rec FingerprintRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return Fingerprint{}, err
	}
	return Fingerprint{Hash: rec.Hash, PortSet: rec.PortSet}, nil
}
