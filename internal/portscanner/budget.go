package portscanner

import "sync"

// Budget enforces a maximum number of notifications dispatched per scan
// cycle, preventing alert storms when many ports change simultaneously.
type Budget struct {
	mu      sync.Mutex
	max     int
	spent   int
}

// NewBudget creates a Budget allowing at most max events per cycle.
// A max of 0 means unlimited.
func NewBudget(max int) *Budget {
	return &Budget{max: max}
}

// Allow returns true and increments the counter if the budget has not
// been exhausted. It always returns true when max is 0.
func (b *Budget) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.max == 0 {
		return true
	}
	if b.spent >= b.max {
		return false
	}
	b.spent++
	return true
}

// Reset zeroes the counter; call this at the start of each scan cycle.
func (b *Budget) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spent = 0
}

// Remaining returns how many events can still be dispatched this cycle.
// Returns -1 when budget is unlimited.
func (b *Budget) Remaining() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.max == 0 {
		return -1
	}
	r := b.max - b.spent
	if r < 0 {
		return 0
	}
	return r
}

// Apply returns the leading slice of events that fit within the budget.
func (b *Budget) Apply(events []ChangeEvent) []ChangeEvent {
	out := make([]ChangeEvent, 0, len(events))
	for _, e := range events {
		if !b.Allow() {
			break
		}
		out = append(out, e)
	}
	return out
}
