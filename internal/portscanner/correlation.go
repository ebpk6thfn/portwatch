package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// CorrelationPolicy controls how events are correlated.
type CorrelationPolicy struct {
	Window   time.Duration
	MinCount int
}

// DefaultCorrelationPolicy returns sensible defaults.
func DefaultCorrelationPolicy() CorrelationPolicy {
	return CorrelationPolicy{
		Window:   30 * time.Second,
		MinCount: 3,
	}
}

// CorrelatedGroup represents a set of related change events.
type CorrelatedGroup struct {
	ID        string
	Events    []ChangeEvent
	FirstSeen time.Time
	LastSeen  time.Time
	Protocol  string
}

func (g CorrelatedGroup) String() string {
	return fmt.Sprintf("group=%s protocol=%s events=%d window=[%s..%s]",
		g.ID, g.Protocol, len(g.Events),
		g.FirstSeen.Format(time.RFC3339),
		g.LastSeen.Format(time.RFC3339))
}

// Correlator groups related ChangeEvents within a sliding time window.
type Correlator struct {
	mu     sync.Mutex
	policy CorrelationPolicy
	buckets map[string]*CorrelatedGroup
	now    func() time.Time
}

// NewCorrelator creates a Correlator with the given policy.
func NewCorrelator(policy CorrelationPolicy, now func() time.Time) *Correlator {
	if now == nil {
		now = time.Now
	}
	return &Correlator{
		policy:  policy,
		buckets: make(map[string]*CorrelatedGroup),
		now:     now,
	}
}

// Add records an event and returns any completed group (nil if not yet ready).
func (c *Correlator) Add(event ChangeEvent) *CorrelatedGroup {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s:%s", event.Entry.Protocol, event.Type)
	t := c.now()

	g, ok := c.buckets[key]
	if !ok || t.Sub(g.FirstSeen) > c.policy.Window {
		g = &CorrelatedGroup{
			ID:        fmt.Sprintf("%s-%d", key, t.UnixNano()),
			FirstSeen: t,
			Protocol:  event.Entry.Protocol,
		}
		c.buckets[key] = g
	}

	g.Events = append(g.Events, event)
	g.LastSeen = t

	if len(g.Events) >= c.policy.MinCount {
		result := *g
		delete(c.buckets, key)
		return &result
	}
	return nil
}

// Flush returns and clears all in-progress groups regardless of MinCount.
func (c *Correlator) Flush() []CorrelatedGroup {
	c.mu.Lock()
	defer c.mu.Unlock()

	var out []CorrelatedGroup
	for key, g := range c.buckets {
		out = append(out, *g)
		delete(c.buckets, key)
	}
	return out
}
