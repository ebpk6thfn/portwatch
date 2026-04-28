package portscanner

import (
	"sync"
	"time"
)

// StaggerPolicy controls how events are staggered across a time window.
type StaggerPolicy struct {
	Window  time.Duration
	MaxSlot int
}

// DefaultStaggerPolicy returns a sensible default stagger policy.
func DefaultStaggerPolicy() StaggerPolicy {
	return StaggerPolicy{
		Window:  5 * time.Second,
		MaxSlot: 10,
	}
}

// Stagger spreads event dispatch across a time window to avoid thundering
// herds when many events arrive simultaneously.
type Stagger struct {
	policy StaggerPolicy
	mu     sync.Mutex
	slots  map[string]int // key -> slot index assigned
	now    func() time.Time
}

// NewStagger creates a new Stagger with the given policy.
func NewStagger(policy StaggerPolicy) *Stagger {
	return newStaggerWithClock(policy, time.Now)
}

func newStaggerWithClock(policy StaggerPolicy, now func() time.Time) *Stagger {
	if policy.MaxSlot <= 0 {
		policy.MaxSlot = DefaultStaggerPolicy().MaxSlot
	}
	if policy.Window <= 0 {
		policy.Window = DefaultStaggerPolicy().Window
	}
	return &Stagger{
		policy: policy,
		slots:  make(map[string]int),
		now:    now,
	}
}

// Delay returns the staggered delay for the given key. Each unique key is
// assigned a deterministic slot within [0, MaxSlot), and the delay is
// proportional to that slot index within the configured window.
func (s *Stagger) Delay(key string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	slot, ok := s.slots[key]
	if !ok {
		slot = len(s.slots) % s.policy.MaxSlot
		s.slots[key] = slot
	}

	slotDuration := s.policy.Window / time.Duration(s.policy.MaxSlot)
	return time.Duration(slot) * slotDuration
}

// Reset clears all assigned slots.
func (s *Stagger) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slots = make(map[string]int)
}

// Len returns the number of tracked keys.
func (s *Stagger) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.slots)
}
