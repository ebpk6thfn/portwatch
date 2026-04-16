package portscanner

import "sync"

// WatchlistAction defines what to do when a watchlist rule matches.
type WatchlistAction string

const (
	ActionAlert  WatchlistAction = "alert"
	ActionIgnore WatchlistAction = "ignore"
)

// WatchlistRule pairs a port number with an action.
type WatchlistRule struct {
	Port   uint16
	Action WatchlistAction
}

// Watchlist holds a set of port rules and evaluates ChangeEvents against them.
type Watchlist struct {
	mu    sync.RWMutex
	rules map[uint16]WatchlistAction
}

// NewWatchlist creates a Watchlist from a slice of WatchlistRules.
func NewWatchlist(rules []WatchlistRule) *Watchlist {
	w := &Watchlist{
		rules: make(map[uint16]WatchlistAction, len(rules)),
	}
	for _, r := range rules {
		w.rules[r.Port] = r.Action
	}
	return w
}

// Add inserts or replaces a rule for the given port.
func (w *Watchlist) Add(port uint16, action WatchlistAction) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.rules[port] = action
}

// Remove deletes any rule associated with the given port.
func (w *Watchlist) Remove(port uint16) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.rules, port)
}

// Evaluate returns the action for the event's port and whether a rule matched.
// If no rule exists the second return value is false.
func (w *Watchlist) Evaluate(event ChangeEvent) (WatchlistAction, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	action, ok := w.rules[event.Entry.Port]
	return action, ok
}

// Filter returns only those events that have a matching rule with ActionAlert,
// plus all events that have no rule at all (pass-through behaviour).
func (w *Watchlist) Filter(events []ChangeEvent) []ChangeEvent {
	out := make([]ChangeEvent, 0, len(events))
	for _, e := range events {
		action, matched := w.Evaluate(e)
		if !matched || action == ActionAlert {
			out = append(out, e)
		}
	}
	return out
}
