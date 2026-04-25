package portscanner

import (
	"sync"
	"time"
)

// LimboPolicy controls how long events wait in limbo before being confirmed or discarded.
type LimboPolicy struct {
	Window    time.Duration
	MaxSize   int
}

// DefaultLimboPolicy returns a sensible default limbo policy.
func DefaultLimboPolicy() LimboPolicy {
	return LimboPolicy{
		Window:  5 * time.Second,
		MaxSize: 256,
	}
}

type limboEntry struct {
	event     ChangeEvent
	arrivedAt time.Time
}

// Limbo holds events in a waiting state until they are either confirmed
// (seen again in the next scan) or expire without confirmation.
// This prevents transient port flaps from generating spurious alerts.
type Limbo struct {
	mu     sync.Mutex
	policy LimboPolicy
	now    func() time.Time
	store  map[string]limboEntry
}

// NewLimbo creates a Limbo with the given policy.
func NewLimbo(policy LimboPolicy, now func() time.Time) *Limbo {
	if now == nil {
		now = time.Now
	}
	return &Limbo{
		policy: policy,
		now:    now,
		store:  make(map[string]limboEntry),
	}
}

// Hold places an event into limbo. Returns true if the event was newly added,
// false if it was already present (i.e. confirmed — caller should emit it).
func (l *Limbo) Hold(ev ChangeEvent) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := ev.Entry.Key()
	if _, exists := l.store[key]; exists {
		delete(l.store, key)
		return false // confirmed — emit the event
	}

	if len(l.store) >= l.policy.MaxSize {
		l.evictOldestLocked()
	}

	l.store[key] = limboEntry{event: ev, arrivedAt: l.now()}
	return true // held — do not emit yet
}

// Flush returns all events that have been in limbo longer than the window
// without confirmation, removing them from the store.
func (l *Limbo) Flush() []ChangeEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	var expired []ChangeEvent
	for key, entry := range l.store {
		if now.Sub(entry.arrivedAt) >= l.policy.Window {
			expired = append(expired, entry.event)
			delete(l.store, key)
		}
	}
	return expired
}

// Len returns the number of events currently in limbo.
func (l *Limbo) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.store)
}

func (l *Limbo) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, e := range l.store {
		if first || e.arrivedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = e.arrivedAt
			first = false
		}
	}
	if oldestKey != "" {
		delete(l.store, oldestKey)
	}
}
