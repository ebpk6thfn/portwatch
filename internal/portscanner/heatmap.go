package portscanner

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// HeatmapEntry records activity intensity for a port over time.
type HeatmapEntry struct {
	Port     uint16
	Protocol string
	Hits     int
	LastSeen time.Time
}

// String returns a human-readable representation of the entry.
func (e HeatmapEntry) String() string {
	return fmt.Sprintf("%s/%d hits=%d last=%s", e.Protocol, e.Port, e.Hits, e.LastSeen.Format(time.RFC3339))
}

// Heatmap tracks how frequently ports appear in change events within a
// sliding time window, allowing callers to identify the most active ports.
type Heatmap struct {
	mu     sync.Mutex
	window time.Duration
	bucket map[string]*heatmapBucket
}

type heatmapBucket struct {
	entry     HeatmapEntry
	timestamp []time.Time
}

// NewHeatmap creates a Heatmap that retains activity within the given window.
func NewHeatmap(window time.Duration) *Heatmap {
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &Heatmap{
		window: window,
		bucket: make(map[string]*heatmapBucket),
	}
}

// Record registers a ChangeEvent hit in the heatmap.
func (h *Heatmap) Record(event ChangeEvent, now time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := fmt.Sprintf("%s/%d", event.Entry.Protocol, event.Entry.Port)
	b, ok := h.bucket[key]
	if !ok {
		b = &heatmapBucket{
			entry: HeatmapEntry{
				Port:     event.Entry.Port,
				Protocol: event.Entry.Protocol,
			},
		}
		h.bucket[key] = b
	}
	b.timestamp = append(b.timestamp, now)
	b.entry.LastSeen = now
	h.evict(b, now)
	b.entry.Hits = len(b.timestamp)
}

// Top returns up to n entries ordered by descending hit count within the window.
func (h *Heatmap) Top(n int, now time.Time) []HeatmapEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	var entries []HeatmapEntry
	for _, b := range h.bucket {
		h.evict(b, now)
		if len(b.timestamp) > 0 {
			b.entry.Hits = len(b.timestamp)
			entries = append(entries, b.entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Hits > entries[j].Hits
	})
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

// Len returns the number of distinct ports currently tracked.
func (h *Heatmap) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.bucket)
}

func (h *Heatmap) evict(b *heatmapBucket, now time.Time) {
	cutoff := now.Add(-h.window)
	i := 0
	for i < len(b.timestamp) && b.timestamp[i].Before(cutoff) {
		i++
	}
	b.timestamp = b.timestamp[i:]
}
