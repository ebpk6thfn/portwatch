package portscanner

import "time"

// Snapshot holds a point-in-time capture of all observed port entries
// along with metadata about when the scan occurred.
type Snapshot struct {
	// Entries is the full list of port entries observed during this scan.
	Entries []Entry
	// CapturedAt is the wall-clock time the scan completed.
	CapturedAt time.Time
	// ScanDuration is how long the underlying scan took.
	ScanDuration time.Duration
}

// NewSnapshot creates a Snapshot from a slice of entries, recording the
// current time and the provided scan duration.
func NewSnapshot(entries []Entry, scanDuration time.Duration) Snapshot {
	return Snapshot{
		Entries:      entries,
		CapturedAt:   time.Now(),
		ScanDuration: scanDuration,
	}
}

// Len returns the number of entries in the snapshot.
func (s Snapshot) Len() int {
	return len(s.Entries)
}

// ToMap converts the snapshot entries into a keyed map for efficient
// lookup and diffing operations.
func (s Snapshot) ToMap() map[string]Entry {
	m := make(map[string]Entry, len(s.Entries))
	for _, e := range s.Entries {
		m[e.Key()] = e
	}
	return m
}

// FilteredEntries returns only the entries that pass the provided Filter.
func (s Snapshot) FilteredEntries(f *Filter) []Entry {
	out := make([]Entry, 0, len(s.Entries))
	for _, e := range s.Entries {
		if f.Allow(e) {
			out = append(out, e)
		}
	}
	return out
}
