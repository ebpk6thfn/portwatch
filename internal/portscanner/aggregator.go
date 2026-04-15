package portscanner

import "time"

// AggregatedEvent groups multiple ChangeEvents that occurred within
// a single scan cycle into a single unit for downstream consumers.
type AggregatedEvent struct {
	// ScannedAt is the time the scan cycle completed.
	ScannedAt time.Time
	// Opened contains entries for ports that were newly opened.
	Opened []Entry
	// Closed contains entries for ports that were recently closed.
	Closed []Entry
}

// IsEmpty reports whether the event contains no changes.
func (a AggregatedEvent) IsEmpty() bool {
	return len(a.Opened) == 0 && len(a.Closed) == 0
}

// TotalChanges returns the total number of port changes in the event.
func (a AggregatedEvent) TotalChanges() int {
	return len(a.Opened) + len(a.Closed)
}

// Aggregator collects ChangeEvents from a scan cycle and produces
// a single AggregatedEvent summarising all changes.
type Aggregator struct {
	now func() time.Time
}

// NewAggregator creates an Aggregator. If now is nil, time.Now is used.
func NewAggregator(now func() time.Time) *Aggregator {
	if now == nil {
		now = time.Now
	}
	return &Aggregator{now: now}
}

// Aggregate converts a slice of ChangeEvents into an AggregatedEvent.
// Events are partitioned into Opened / Closed buckets.
func (a *Aggregator) Aggregate(events []ChangeEvent) AggregatedEvent {
	agg := AggregatedEvent{ScannedAt: a.now()}
	for _, ev := range events {
		switch ev.Type {
		case EventOpened:
			agg.Opened = append(agg.Opened, ev.Entry)
		case EventClosed:
			agg.Closed = append(agg.Closed, ev.Entry)
		}
	}
	return agg
}
