package portscanner

import (
	"encoding/json"
	"os"
	"time"
)

// Baseline represents a saved snapshot used as a reference point
// to suppress alerts for ports that were already open at startup.
type Baseline struct {
	CreatedAt time.Time        `json:"created_at"`
	Entries   map[string]bool  `json:"entries"`
}

// NewBaseline creates a Baseline from the given entries.
func NewBaseline(entries []Entry) *Baseline {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		m[e.Key()] = true
	}
	return &Baseline{
		CreatedAt: time.Now(),
		Entries:   m,
	}
}

// Contains returns true if the entry key exists in the baseline.
func (b *Baseline) Contains(e Entry) bool {
	return b.Entries[e.Key()]
}

// SaveBaseline writes the baseline to the given path as JSON.
func SaveBaseline(path string, b *Baseline) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(b)
}

// LoadBaseline reads a baseline from the given path.
func LoadBaseline(path string) (*Baseline, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Baseline{Entries: make(map[string]bool)}, nil
		}
		return nil, err
	}
	defer f.Close()
	var b Baseline
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return nil, err
	}
	if b.Entries == nil {
		b.Entries = make(map[string]bool)
	}
	return &b, nil
}
