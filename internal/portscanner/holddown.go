package portscanner

import (
	"sync"
	"time"
)

// HolddownPolicy configures the hold-down timer behaviour.
type HolddownPolicy struct {
	// Duration is how long a key must be stable (absent) before a
	// "port closed" event is emitted. Zero disables hold-down.
	Duration time.Duration
}

// DefaultHolddownPolicy returns a sensible default.
func DefaultHolddownPolicy() HolddownPolicy {
	return HolddownPolicy{Duration: 10 * time.Second}
}

// holddownEntry tracks when a key first went absent.
type holddownEntry struct {
	firstAbsent time.Time
}

// Holddown suppresses transient "port closed" events by requiring a
// port to remain absent for at least Policy.Duration before the close
// event is forwarded downstream.
type Holddown struct {
	mu      sync.Mutex
	policy  HolddownPolicy
	pending map[string]holddownEntry
	nowFn   func() time.Time
}

// NewHolddown creates a Holddown with the given policy.
func NewHolddown(policy HolddownPolicy) *Holddown {
	return &Holddown{
		policy:  policy,
		pending: make(map[string]holddownEntry),
		nowFn:   time.Now,
	}
}

// Evaluate processes a slice of ChangeEvents and returns only those
// that should be forwarded. Opened events pass through immediately.
// Closed events are held until the key has been absent for the full
// hold-down duration.
func (h *Holddown) Evaluate(events []ChangeEvent) []ChangeEvent {
	if h.policy.Duration == 0 {
		return events
	}

	now := h.nowFn()
	h.mu.Lock()
	defer h.mu.Unlock()

	// Track which keys are currently "open" so we can clear them.
	openKeys := make(map[string]struct{})
	var out []ChangeEvent

	for _, ev := range events {
		key := ev.Entry.Key()
		switch ev.Type {
		case EventOpened:
			// Port re-appeared — cancel any pending hold-down.
			delete(h.pending, key)
			openKeys[key] = struct{}{}
			out = append(out, ev)
		case EventClosed:
			if _, ok := h.pending[key]; !ok {
				h.pending[key] = holddownEntry{firstAbsent: now}
			}
			entry := h.pending[key]
			if now.Sub(entry.firstAbsent) >= h.policy.Duration {
				delete(h.pending, key)
				out = append(out, ev)
			}
			// else: still within hold-down window — drop for now.
		}
	}

	// Purge keys that have re-opened (already handled above).
	for k := range openKeys {
		delete(h.pending, k)
	}

	return out
}

// Flush forces all pending hold-down entries to be emitted regardless
// of elapsed time. Useful on graceful shutdown.
func (h *Holddown) Flush(buildEvent func(key string) ChangeEvent) []ChangeEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	var out []ChangeEvent
	for key := range h.pending {
		out = append(out, buildEvent(key))
	}
	h.pending = make(map[string]holddownEntry)
	return out
}

// PendingCount returns the number of events currently held.
func (h *Holddown) PendingCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.pending)
}
