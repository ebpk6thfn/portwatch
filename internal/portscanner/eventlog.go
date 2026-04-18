package portscanner

import (
	"sync"
	"time"
)

// EventLogEntry records a single emitted change event with a timestamp.
type EventLogEntry struct {
	Timestamp time.Time
	Event     ChangeEvent
}

// EventLog is a bounded, thread-safe log of recent ChangeEvents.
type EventLog struct {
	mu      sync.Mutex
	entries []EventLogEntry
	maxSize int
}

// NewEventLog creates an EventLog that retains at most maxSize entries.
func NewEventLog(maxSize int) *EventLog {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &EventLog{maxSize: maxSize}
}

// Record appends a ChangeEvent to the log, evicting the oldest if full.
func (l *EventLog) Record(ev ChangeEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := EventLogEntry{Timestamp: time.Now(), Event: ev}
	if len(l.entries) >= l.maxSize {
		l.entries = append(l.entries[1:], entry)
	} else {
		l.entries = append(l.entries, entry)
	}
}

// All returns a copy of all log entries in insertion order.
func (l *EventLog) All() []EventLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]EventLogEntry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Len returns the number of entries currently stored.
func (l *EventLog) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// Clear removes all entries from the log.
func (l *EventLog) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}

// Since returns all entries recorded at or after the given time.
func (l *EventLog) Since(t time.Time) []EventLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []EventLogEntry
	for _, e := range l.entries {
		if !e.Timestamp.Before(t) {
			out = append(out, e)
		}
	}
	return out
}
