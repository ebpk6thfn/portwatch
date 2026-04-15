package portscanner

// History is a bounded ring-buffer of Snapshots.
type History struct {
	snaps []*Snapshot
	cap   int
}

// NewHistory creates a History that retains at most maxLen snapshots.
func NewHistory(maxLen int) *History {
	if maxLen <= 0 {
		maxLen = 10
	}
	return &History{cap: maxLen}
}

// Add appends a new snapshot, evicting the oldest when the buffer is full.
func (h *History) Add(s *Snapshot) {
	if len(h.snaps) >= h.cap {
		h.snaps = h.snaps[1:]
	}
	h.snaps = append(h.snaps, s)
}

// Latest returns the most recently added snapshot, or nil if empty.
func (h *History) Latest() *Snapshot {
	if len(h.snaps) == 0 {
		return nil
	}
	return h.snaps[len(h.snaps)-1]
}

// Previous returns the snapshot immediately before the latest, or nil when
// fewer than two snapshots have been recorded.
func (h *History) Previous() *Snapshot {
	if len(h.snaps) < 2 {
		return nil
	}
	return h.snaps[len(h.snaps)-2]
}

// Len returns the number of snapshots currently held.
func (h *History) Len() int { return len(h.snaps) }

// All returns a slice of all retained snapshots in chronological order.
func (h *History) All() []*Snapshot {
	out := make([]*Snapshot, len(h.snaps))
	copy(out, h.snaps)
	return out
}
