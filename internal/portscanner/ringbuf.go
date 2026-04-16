package portscanner

import "sync"

// RingBuffer is a fixed-capacity circular buffer of ChangeEvents.
// Once full, the oldest entry is overwritten.
type RingBuffer struct {
	mu   sync.Mutex
	buf  []ChangeEvent
	cap  int
	head int // next write position
	len  int
}

// NewRingBuffer creates a RingBuffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &RingBuffer{
		buf: make([]ChangeEvent, capacity),
		cap: capacity,
	}
}

// Push adds an event to the buffer, overwriting the oldest if full.
func (r *RingBuffer) Push(e ChangeEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = e
	r.head = (r.head + 1) % r.cap
	if r.len < r.cap {
		r.len++
	}
}

// Drain returns all buffered events in insertion order and clears the buffer.
func (r *RingBuffer) Drain() []ChangeEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.len == 0 {
		return nil
	}
	out := make([]ChangeEvent, r.len)
	start := (r.head - r.len + r.cap) % r.cap
	for i := 0; i < r.len; i++ {
		out[i] = r.buf[(start+i)%r.cap]
	}
	r.head = 0
	r.len = 0
	return out
}

// Len returns the current number of buffered events.
func (r *RingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.len
}

// Cap returns the maximum capacity of the buffer.
func (r *RingBuffer) Cap() int { return r.cap }
