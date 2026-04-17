package portscanner

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ScoreEntry holds the aggregated score and hit count for a single key.
type ScoreEntry struct {
	Key       string
	Score     float64
	Hits      int
	LastSeen  time.Time
}

func (e ScoreEntry) String() string {
	return fmt.Sprintf("%s score=%.2f hits=%d", e.Key, e.Score, e.Hits)
}

// Scoreboard tracks per-key scores over time, evicting entries older than ttl.
type Scoreboard struct {
	mu      sync.Mutex
	entries map[string]*ScoreEntry
	ttl     time.Duration
	now     func() time.Time
}

// NewScoreboard creates a Scoreboard with the given TTL.
func NewScoreboard(ttl time.Duration) *Scoreboard {
	return &Scoreboard{
		entries: make(map[string]*ScoreEntry),
		ttl:     ttl,
		now:     time.Now,
	}
}

// Add increments the score and hit count for key.
func (s *Scoreboard) Add(key string, delta float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evict()
	e, ok := s.entries[key]
	if !ok {
		e = &ScoreEntry{Key: key}
		s.entries[key] = e
	}
	e.Score += delta
	e.Hits++
	e.LastSeen = s.now()
}

// Top returns the top n entries sorted by score descending.
func (s *Scoreboard) Top(n int) []ScoreEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evict()
	all := make([]ScoreEntry, 0, len(s.entries))
	for _, e := range s.entries {
		all = append(all, *e)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Score > all[j].Score })
	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

// Len returns the number of active entries.
func (s *Scoreboard) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evict()
	return len(s.entries)
}

func (s *Scoreboard) evict() {
	cutoff := s.now().Add(-s.ttl)
	for k, e := range s.entries {
		if e.LastSeen.Before(cutoff) {
			delete(s.entries, k)
		}
	}
}
