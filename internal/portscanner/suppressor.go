package portscanner

import (
	"sync"
	"time"
)

// Suppressor prevents repeated notifications for the same port event
// within a configurable quiet period after the first alert.
type Suppressor struct {
	mu        sync.Mutex
	quietFor  time.Duration
	suppressed map[string]time.Time
	now       func() time.Time
}

// NewSuppressor creates a Suppressor that silences repeated events
// for the same key within quietFor duration.
func NewSuppressor(quietFor time.Duration) *Suppressor {
	return &Suppressor{
		quietFor:   quietFor,
		suppressed: make(map[string]time.Time),
		now:        time.Now,
	}
}

// IsSuppressed returns true if the event key has been seen recently
// and should be silenced. It records the key on first encounter.
func (s *Suppressor) IsSuppressed(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if last, ok := s.suppressed[key]; ok {
		if now.Sub(last) < s.quietFor {
			return true
		}
	}
	s.suppressed[key] = now
	return false
}

// Flush removes all entries whose quiet period has expired.
func (s *Suppressor) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for k, last := range s.suppressed {
		if now.Sub(last) >= s.quietFor {
			delete(s.suppressed, k)
		}
	}
}

// Filter returns only the events that are not currently suppressed.
func (s *Suppressor) Filter(events []ChangeEvent) []ChangeEvent {
	out := make([]ChangeEvent, 0, len(events))
	for _, e := range events {
		if !s.IsSuppressed(dedupKey(e)) {
			out = append(out, e)
		}
	}
	return out
}
