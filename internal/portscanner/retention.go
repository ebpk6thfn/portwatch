package portscanner

import (
	"sync"
	"time"
)

// RetentionPolicy defines how long events are kept.
type RetentionPolicy struct {
	MaxAge time.Duration
	MaxCount int
}

// RetentionStore stores ChangeEvents and evicts them based on a policy.
type RetentionStore struct {
	mu     sync.Mutex
	events []timedEvent
	policy RetentionPolicy
	now    func() time.Time
}

type timedEvent struct {
	at    time.Time
	event ChangeEvent
}

// NewRetentionStore creates a RetentionStore with the given policy.
func NewRetentionStore(policy RetentionPolicy) *RetentionStore {
	return &RetentionStore{
		policy: policy,
		now:    time.Now,
	}
}

// Add records a new event, then evicts stale or excess entries.
func (r *RetentionStore) Add(e ChangeEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, timedEvent{at: r.now(), event: e})
	r.evict()
}

// All returns all retained events.
func (r *RetentionStore) All() []ChangeEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evict()
	out := make([]ChangeEvent, len(r.events))
	for i, te := range r.events {
		out[i] = te.event
	}
	return out
}

// Len returns the number of retained events.
func (r *RetentionStore) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evict()
	return len(r.events)
}

// evict removes events that are too old or exceed MaxCount. Must be called with lock held.
func (r *RetentionStore) evict() {
	if r.policy.MaxAge > 0 {
		cutoff := r.now().Add(-r.policy.MaxAge)
		start := 0
		for start < len(r.events) && r.events[start].at.Before(cutoff) {
			start++
		}
		r.events = r.events[start:]
	}
	if r.policy.MaxCount > 0 && len(r.events) > r.policy.MaxCount {
		r.events = r.events[len(r.events)-r.policy.MaxCount:]
	}
}
