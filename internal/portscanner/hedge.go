package portscanner

import (
	"sync"
	"time"
)

// HedgePolicy controls how the hedge filter behaves.
type HedgePolicy struct {
	// Window is how long to wait before confirming a change is real.
	Window time.Duration
	// MaxPending is the maximum number of pending events held in the hedge.
	MaxPending int
}

// DefaultHedgePolicy returns a sensible default HedgePolicy.
func DefaultHedgePolicy() HedgePolicy {
	return HedgePolicy{
		Window:     3 * time.Second,
		MaxPending: 256,
	}
}

type hedgeEntry struct {
	event     ChangeEvent
	receivedAt time.Time
}

// Hedge holds events for a short window and only emits them if they are not
// immediately reversed (e.g. a port that opens and closes within the window).
type Hedge struct {
	mu      sync.Mutex
	policy  HedgePolicy
	pending map[string]hedgeEntry
	now     func() time.Time
}

// NewHedge creates a Hedge with the given policy.
func NewHedge(policy HedgePolicy) *Hedge {
	return newHedgeWithClock(policy, time.Now)
}

func newHedgeWithClock(policy HedgePolicy, now func() time.Time) *Hedge {
	if policy.MaxPending <= 0 {
		policy.MaxPending = DefaultHedgePolicy().MaxPending
	}
	if policy.Window <= 0 {
		policy.Window = DefaultHedgePolicy().Window
	}
	return &Hedge{
		policy:  policy,
		pending: make(map[string]hedgeEntry),
		now:     now,
	}
}

// Hold registers an event. Returns true if the event was newly admitted to the
// pending set, false if it was cancelled by an opposing event.
func (h *Hedge) Hold(event ChangeEvent) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := event.Entry.Key()
	if existing, ok := h.pending[key]; ok {
		// Opposing direction — cancel both.
		if existing.event.Type != event.Type {
			delete(h.pending, key)
			return false
		}
	}
	if len(h.pending) >= h.policy.MaxPending {
		// Evict oldest to make room.
		h.evictOldest()
	}
	h.pending[key] = hedgeEntry{event: event, receivedAt: h.now()}
	return true
}

// Flush returns all events whose hedge window has expired and removes them.
func (h *Hedge) Flush() []ChangeEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.now()
	var out []ChangeEvent
	for key, entry := range h.pending {
		if now.Sub(entry.receivedAt) >= h.policy.Window {
			out = append(out, entry.event)
			delete(h.pending, key)
		}
	}
	return out
}

// Len returns the number of currently pending events.
func (h *Hedge) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.pending)
}

func (h *Hedge) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range h.pending {
		if oldestKey == "" || entry.receivedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.receivedAt
		}
	}
	if oldestKey != "" {
		delete(h.pending, oldestKey)
	}
}
