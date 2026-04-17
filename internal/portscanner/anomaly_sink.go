package portscanner

import (
	"sync"
	"time"
)

// AnomalySink collects anomalies produced during a scan cycle and
// exposes them for consumption by the daemon or notifiers.
type AnomalySink struct {
	mu       sync.Mutex
	items    []Anomaly
	maxSize  int
}

// NewAnomalySink creates a sink that retains at most maxSize anomalies.
func NewAnomalySink(maxSize int) *AnomalySink {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &AnomalySink{maxSize: maxSize}
}

// Push adds an anomaly to the sink, dropping the oldest if at capacity.
func (s *AnomalySink) Push(a Anomaly) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) >= s.maxSize {
		s.items = s.items[1:]
	}
	s.items = append(s.items, a)
}

// Drain returns all buffered anomalies and clears the sink.
func (s *AnomalySink) Drain() []Anomaly {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Anomaly, len(s.items))
	copy(out, s.items)
	s.items = s.items[:0]
	return out
}

// Len returns the current number of buffered anomalies.
func (s *AnomalySink) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

// ProcessEvents runs each event through the detector and pushes any anomaly.
func (s *AnomalySink) ProcessEvents(detector *AnomalyDetector, events []ChangeEvent, now time.Time) {
	for _, ev := range events {
		if a := detector.Evaluate(ev, now); a != nil {
			s.Push(*a)
		}
	}
}
