package portscanner

import (
	"sync"
	"time"
)

// DrainPolicy controls how a Drainer flushes buffered events.
type DrainPolicy struct {
	MaxBuffer int
	MaxAge    time.Duration
}

// DefaultDrainPolicy returns sensible defaults.
func DefaultDrainPolicy() DrainPolicy {
	return DrainPolicy{
		MaxBuffer: 64,
		MaxAge:    10 * time.Second,
	}
}

// drainEntry holds an event and its arrival time.
type drainEntry struct {
	event ChangeEvent
	at    time.Time
}

// Drainer buffers ChangeEvents and flushes them when the buffer is full
// or when the oldest entry exceeds MaxAge.
type Drainer struct {
	mu     sync.Mutex
	policy DrainPolicy
	buf    []drainEntry
	now    func() time.Time
}

// NewDrainer creates a Drainer with the given policy.
func NewDrainer(p DrainPolicy) *Drainer {
	if p.MaxBuffer <= 0 {
		p.MaxBuffer = DefaultDrainPolicy().MaxBuffer
	}
	if p.MaxAge <= 0 {
		p.MaxAge = DefaultDrainPolicy().MaxAge
	}
	return &Drainer{policy: p, now: time.Now}
}

// Push adds an event to the buffer. Returns flushed events if threshold met.
func (d *Drainer) Push(e ChangeEvent) []ChangeEvent {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.buf = append(d.buf, drainEntry{event: e, at: d.now()})
	if len(d.buf) >= d.policy.MaxBuffer {
		return d.flush()
	}
	return nil
}

// Tick checks age-based flushing; call periodically.
func (d *Drainer) Tick() []ChangeEvent {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.buf) == 0 {
		return nil
	}
	if d.now().Sub(d.buf[0].at) >= d.policy.MaxAge {
		return d.flush()
	}
	return nil
}

// flush drains the buffer and returns all events. Must be called with lock held.
func (d *Drainer) flush() []ChangeEvent {
	out := make([]ChangeEvent, len(d.buf))
	for i, e := range d.buf {
		out[i] = e.event
	}
	d.buf = d.buf[:0]
	return out
}

// Len returns the current buffer length.
func (d *Drainer) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.buf)
}
