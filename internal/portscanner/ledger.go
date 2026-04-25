package portscanner

import (
	"sync"
	"time"
)

// LedgerEntry records the first and last time a port event was observed.
type LedgerEntry struct {
	Key       string
	Port      int
	Protocol  string
	Process   string
	FirstSeen time.Time
	LastSeen  time.Time
	Count     int
}

// Ledger maintains a running record of all observed port change events,
// tracking first/last seen timestamps and occurrence counts per unique key.
type Ledger struct {
	mu      sync.RWMutex
	entries map[string]*LedgerEntry
	maxSize int
}

// NewLedger creates a Ledger with the given maximum number of tracked entries.
// A maxSize of 0 means unlimited.
func NewLedger(maxSize int) *Ledger {
	return &Ledger{
		entries: make(map[string]*LedgerEntry),
		maxSize: maxSize,
	}
}

// Record adds or updates a ledger entry for the given event.
func (l *Ledger) Record(e ChangeEvent, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := e.Entry.Key()
	if existing, ok := l.entries[key]; ok {
		existing.LastSeen = now
		existing.Count++
		return
	}

	if l.maxSize > 0 && len(l.entries) >= l.maxSize {
		l.evictOldest()
	}

	l.entries[key] = &LedgerEntry{
		Key:       key,
		Port:      e.Entry.Port,
		Protocol:  e.Entry.Protocol,
		Process:   e.Entry.Process,
		FirstSeen: now,
		LastSeen:  now,
		Count:     1,
	}
}

// Get returns the ledger entry for the given key, or false if not found.
func (l *Ledger) Get(key string) (LedgerEntry, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	e, ok := l.entries[key]
	if !ok {
		return LedgerEntry{}, false
	}
	return *e, true
}

// All returns a snapshot of all current ledger entries.
func (l *Ledger) All() []LedgerEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]LedgerEntry, 0, len(l.entries))
	for _, e := range l.entries {
		out = append(out, *e)
	}
	return out
}

// Len returns the number of tracked entries.
func (l *Ledger) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// evictOldest removes the entry with the oldest FirstSeen timestamp.
// Caller must hold the write lock.
func (l *Ledger) evictOldest() {
	var oldest string
	var oldestTime time.Time
	for k, e := range l.entries {
		if oldest == "" || e.FirstSeen.Before(oldestTime) {
			oldest = k
			oldestTime = e.FirstSeen
		}
	}
	if oldest != "" {
		delete(l.entries, oldest)
	}
}
