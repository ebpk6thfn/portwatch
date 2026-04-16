package portscanner

import (
	"sort"
	"time"
)

// ReplayEvent represents a historical change event with a timestamp for replay.
type ReplayEvent struct {
	At    time.Time
	Event ChangeEvent
}

// Replayer replays a sequence of historical ChangeEvents in chronological order.
type Replayer struct {
	events []ReplayEvent
}

// NewReplayer creates a Replayer from a slice of ReplayEvents.
func NewReplayer(events []ReplayEvent) *Replayer {
	sorted := make([]ReplayEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].At.Before(sorted[j].At)
	})
	return &Replayer{events: sorted}
}

// All returns all replay events in chronological order.
func (r *Replayer) All() []ReplayEvent {
	return r.events
}

// Between returns events whose timestamp falls within [from, to].
func (r *Replayer) Between(from, to time.Time) []ReplayEvent {
	var out []ReplayEvent
	for _, e := range r.events {
		if !e.At.Before(from) && !e.At.After(to) {
			out = append(out, e)
		}
	}
	return out
}

// Since returns events at or after the given time.
func (r *Replayer) Since(t time.Time) []ReplayEvent {
	var out []ReplayEvent
	for _, e := range r.events {
		if !e.At.Before(t) {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the total number of stored events.
func (r *Replayer) Len() int {
	return len(r.events)
}
