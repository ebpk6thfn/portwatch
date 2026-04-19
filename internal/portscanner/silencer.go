package portscanner

import (
	"sync"
	"time"
)

// Silencer suppresses events for specific ports during a configured time window
// (e.g. a maintenance window). Events matching a silenced port are dropped.
type Silencer struct {
	mu      sync.Mutex
	rules   []silenceRule
	nowFunc func() time.Time
}

type silenceRule struct {
	port  uint16
	until time.Time
}

// NewSilencer creates a Silencer with an optional clock override.
func NewSilencer(nowFunc func() time.Time) *Silencer {
	if nowFunc == nil {
		nowFunc = time.Now
	}
	return &Silencer{nowFunc: nowFunc}
}

// Silence suppresses events for port until the given time.
func (s *Silencer) Silence(port uint16, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, silenceRule{port: port, until: until})
}

// IsSilenced returns true if the port is currently suppressed.
func (s *Silencer) IsSilenced(port uint16) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFunc()
	for _, r := range s.rules {
		if r.port == port && now.Before(r.until) {
			return true
		}
	}
	return false
}

// Filter returns only events whose port is not currently silenced.
func (s *Silencer) Filter(events []ChangeEvent) []ChangeEvent {
	out := events[:0:0]
	for _, e := range events {
		if !s.IsSilenced(e.Entry.Port) {
			out = append(out, e)
		}
	}
	return out
}

// Flush removes expired silence rules.
func (s *Silencer) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFunc()
	kept := s.rules[:0]
	for _, r := range s.rules {
		if now.Before(r.until) {
			kept = append(kept, r)
		}
	}
	s.rules = kept
}
