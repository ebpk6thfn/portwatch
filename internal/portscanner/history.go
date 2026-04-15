package portscanner

import "sync"

// History maintains a bounded ring buffer of recent Snapshots, allowing
// callers to inspect recent scan results without persisting them to disk.
type History struct {
	mu       sync.RWMutex
	buf      []Snapshot
	cap      int
	head     int
	count    int
}

// NewHistory creates a History that retains at most maxSnapshots entries.
// maxSnapshots must be >= 1; if it is less, it defaults to 1.
func NewHistory(maxSnapshots int) *History {
	if maxSnapshots < 1 {
		maxSnapshots = 1
	}
	return &History{
		buf: make([]Snapshot, maxSnapshots),
		cap: maxSnapshots,
	}
}

// Add appends a snapshot to the history, evicting the oldest if the buffer
// is full.
func (h *History) Add(s Snapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buf[h.head] = s
	h.head = (h.head + 1) % h.cap
	if h.count < h.cap {
		h.count++
	}
}

// Latest returns the most recently added Snapshot and true, or a zero
// Snapshot and false if history is empty.
func (h *History) Latest() (Snapshot, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.count == 0 {
		return Snapshot{}, false
	}
	idx := (h.head - 1 + h.cap) % h.cap
	return h.buf[idx], true
}

// All returns snapshots in chronological order (oldest first).
func (h *History) All() []Snapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Snapshot, h.count)
	for i := 0; i < h.count; i++ {
		out[i] = h.buf[(h.head-h.count+i+h.cap)%h.cap]
	}
	return out
}

// Len returns the number of snapshots currently stored.
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}
