package portscanner

import "time"

// Snapshot captures a point-in-time view of the active port entries returned
// by a single scan cycle.
type Snapshot struct {
	// entries holds the raw (filtered) entries for this snapshot.
	entries []Entry
	// taken is the wall-clock time at which the snapshot was created.
	taken time.Time
}

// NewSnapshot constructs a Snapshot from a slice of filtered entries.
func NewSnapshot(entries []Entry, taken time.Time) *Snapshot {
	copied := make([]Entry, len(entries))
	copy(copied, entries)
	return &Snapshot{entries: copied, taken: taken}
}

// Taken returns the timestamp of the snapshot.
func (s *Snapshot) Taken() time.Time { return s.taken }

// Entries returns a copy of the entries stored in the snapshot.
func (s *Snapshot) Entries() []Entry {
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// ToMap converts the snapshot entries into a key→Entry map suitable for
// passing to Diff.
func (s *Snapshot) ToMap() map[string]Entry {
	m := make(map[string]Entry, len(s.entries))
	for _, e := range s.entries {
		m[e.Key()] = e
	}
	return m
}

// FilteredEntries returns only those entries that satisfy the provided
// predicate, without mutating the snapshot.
func (s *Snapshot) FilteredEntries(pred func(Entry) bool) []Entry {
	var out []Entry
	for _, e := range s.entries {
		if pred(e) {
			out = append(out, e)
		}
	}
	return out
}
