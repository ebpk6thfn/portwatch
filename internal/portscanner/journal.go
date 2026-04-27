package portscanner

import (
	"sync"
	"time"
)

// JournalEntry records a single audit entry for a change event.
type JournalEntry struct {
	Timestamp time.Time
	EventKey  string
	Kind      string // "opened" | "closed"
	Protocol  string
	Port      uint16
	Process   string
	Severity  string
	Note      string
}

// JournalPolicy controls retention of journal entries.
type JournalPolicy struct {
	MaxEntries int
}

// DefaultJournalPolicy returns a sensible default.
func DefaultJournalPolicy() JournalPolicy {
	return JournalPolicy{MaxEntries: 500}
}

// Journal is a bounded, thread-safe append-only log of change events.
type Journal struct {
	mu      sync.Mutex
	entries []JournalEntry
	policy  JournalPolicy
}

// NewJournal creates a Journal with the given policy.
func NewJournal(policy JournalPolicy) *Journal {
	if policy.MaxEntries <= 0 {
		policy.MaxEntries = DefaultJournalPolicy().MaxEntries
	}
	return &Journal{
		policy:  policy,
		entries: make([]JournalEntry, 0, policy.MaxEntries),
	}
}

// Record appends an entry, evicting the oldest if at capacity.
func (j *Journal) Record(e JournalEntry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if len(j.entries) >= j.policy.MaxEntries {
		j.entries = j.entries[1:]
	}
	j.entries = append(j.entries, e)
}

// All returns a snapshot of all current journal entries.
func (j *Journal) All() []JournalEntry {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]JournalEntry, len(j.entries))
	copy(out, j.entries)
	return out
}

// Len returns the number of entries currently stored.
func (j *Journal) Len() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return len(j.entries)
}

// Clear removes all entries from the journal.
func (j *Journal) Clear() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.entries = j.entries[:0]
}

// Since returns all entries with a timestamp at or after t.
func (j *Journal) Since(t time.Time) []JournalEntry {
	j.mu.Lock()
	defer j.mu.Unlock()
	var out []JournalEntry
	for _, e := range j.entries {
		if !e.Timestamp.Before(t) {
			out = append(out, e)
		}
	}
	return out
}
