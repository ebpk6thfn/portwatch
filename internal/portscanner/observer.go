package portscanner

import (
	"sync"
	"time"
)

// ObserverEvent is emitted when a tracked port's state is observed repeatedly.
type ObserverEvent struct {
	Entry     Entry
	FirstSeen time.Time
	LastSeen  time.Time
	SeenCount int
}

// ObserverPolicy controls how the Observer behaves.
type ObserverPolicy struct {
	MinObservations int
	Window          time.Duration
}

// DefaultObserverPolicy returns a sensible default policy.
func DefaultObserverPolicy() ObserverPolicy {
	return ObserverPolicy{
		MinObservations: 3,
		Window:          5 * time.Minute,
	}
}

type observerRecord struct {
	firstSeen time.Time
	lastSeen  time.Time
	count     int
}

// Observer tracks how many times a port entry has been observed within a
// rolling window. Once the minimum observation count is reached, it emits
// an ObserverEvent.
type Observer struct {
	mu      sync.Mutex
	policy  ObserverPolicy
	records map[string]*observerRecord
	now     func() time.Time
}

// NewObserver creates an Observer with the given policy.
func NewObserver(policy ObserverPolicy) *Observer {
	return &Observer{
		policy:  policy,
		records: make(map[string]*observerRecord),
		now:     time.Now,
	}
}

// Record registers an observation for the given entry. It returns an
// ObserverEvent and true when the minimum observation threshold is met,
// otherwise nil and false.
func (o *Observer) Record(e Entry) (*ObserverEvent, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := o.now()
	key := e.Key()

	rec, ok := o.records[key]
	if !ok || now.Sub(rec.firstSeen) > o.policy.Window {
		rec = &observerRecord{firstSeen: now}
		o.records[key] = rec
	}

	rec.lastSeen = now
	rec.count++

	if rec.count >= o.policy.MinObservations {
		event := &ObserverEvent{
			Entry:     e,
			FirstSeen: rec.firstSeen,
			LastSeen:  rec.lastSeen,
			SeenCount: rec.count,
		}
		return event, true
	}
	return nil, false
}

// Flush removes all records whose window has expired.
func (o *Observer) Flush() {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := o.now()
	for key, rec := range o.records {
		if now.Sub(rec.firstSeen) > o.policy.Window {
			delete(o.records, key)
		}
	}
}

// Len returns the number of currently tracked entries.
func (o *Observer) Len() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.records)
}
