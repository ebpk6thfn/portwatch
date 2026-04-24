package portscanner

import (
	"sync"
	"time"
)

// DeadLetterReason describes why an event was dead-lettered.
type DeadLetterReason string

const (
	ReasonQuotaExceeded   DeadLetterReason = "quota_exceeded"
	ReasonCircuitOpen     DeadLetterReason = "circuit_open"
	ReasonSuppressed      DeadLetterReason = "suppressed"
	ReasonDeliveryFailed  DeadLetterReason = "delivery_failed"
)

// DeadLetter wraps a ChangeEvent with metadata about why it was rejected.
type DeadLetter struct {
	Event     ChangeEvent
	Reason    DeadLetterReason
	OccurredAt time.Time
}

// DeadLetterQueue stores events that could not be delivered or were dropped.
type DeadLetterQueue struct {
	mu      sync.Mutex
	items   []DeadLetter
	maxSize int
}

// NewDeadLetterQueue creates a DeadLetterQueue with the given capacity.
// When full, the oldest entry is evicted.
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &DeadLetterQueue{maxSize: maxSize}
}

// Push adds an event to the dead-letter queue.
func (q *DeadLetterQueue) Push(event ChangeEvent, reason DeadLetterReason) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) >= q.maxSize {
		q.items = q.items[1:]
	}
	q.items = append(q.items, DeadLetter{
		Event:      event,
		Reason:     reason,
		OccurredAt: time.Now(),
	})
}

// Drain returns all dead-letter entries and clears the queue.
func (q *DeadLetterQueue) Drain() []DeadLetter {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]DeadLetter, len(q.items))
	copy(out, q.items)
	q.items = q.items[:0]
	return out
}

// Len returns the current number of dead-letter entries.
func (q *DeadLetterQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// CountByReason returns a map of reason -> count across all queued entries.
func (q *DeadLetterQueue) CountByReason() map[DeadLetterReason]int {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make(map[DeadLetterReason]int)
	for _, dl := range q.items {
		out[dl.Reason]++
	}
	return out
}
