package portscanner

import (
	"testing"
	"time"
)

func fixedNow() func() time.Time {
	t := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func makeAggEntry(proto, ip string, port uint16) Entry {
	return Entry{Protocol: proto, LocalIP: ip, LocalPort: port}
}

func TestAggregator_EmptyEvents(t *testing.T) {
	agg := NewAggregator(fixedNow())
	result := agg.Aggregate(nil)
	if !result.IsEmpty() {
		t.Errorf("expected empty aggregated event, got %+v", result)
	}
	if result.TotalChanges() != 0 {
		t.Errorf("expected 0 total changes, got %d", result.TotalChanges())
	}
}

func TestAggregator_OnlyOpened(t *testing.T) {
	agg := NewAggregator(fixedNow())
	events := []ChangeEvent{
		{Type: EventOpened, Entry: makeAggEntry("tcp", "0.0.0.0", 8080)},
		{Type: EventOpened, Entry: makeAggEntry("tcp", "0.0.0.0", 9090)},
	}
	result := agg.Aggregate(events)
	if len(result.Opened) != 2 {
		t.Errorf("expected 2 opened, got %d", len(result.Opened))
	}
	if len(result.Closed) != 0 {
		t.Errorf("expected 0 closed, got %d", len(result.Closed))
	}
	if result.TotalChanges() != 2 {
		t.Errorf("expected 2 total changes, got %d", result.TotalChanges())
	}
}

func TestAggregator_OnlyClosed(t *testing.T) {
	agg := NewAggregator(fixedNow())
	events := []ChangeEvent{
		{Type: EventClosed, Entry: makeAggEntry("udp", "127.0.0.1", 5353)},
	}
	result := agg.Aggregate(events)
	if len(result.Closed) != 1 {
		t.Errorf("expected 1 closed, got %d", len(result.Closed))
	}
	if !result.IsEmpty() == false {
		t.Error("expected non-empty event")
	}
}

func TestAggregator_MixedEvents(t *testing.T) {
	agg := NewAggregator(fixedNow())
	events := []ChangeEvent{
		{Type: EventOpened, Entry: makeAggEntry("tcp", "0.0.0.0", 443)},
		{Type: EventClosed, Entry: makeAggEntry("tcp", "0.0.0.0", 80)},
		{Type: EventOpened, Entry: makeAggEntry("udp", "0.0.0.0", 53)},
	}
	result := agg.Aggregate(events)
	if len(result.Opened) != 2 {
		t.Errorf("expected 2 opened, got %d", len(result.Opened))
	}
	if len(result.Closed) != 1 {
		t.Errorf("expected 1 closed, got %d", len(result.Closed))
	}
	if result.TotalChanges() != 3 {
		t.Errorf("expected 3 total changes, got %d", result.TotalChanges())
	}
}

func TestAggregator_ScannedAt(t *testing.T) {
	expected := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	agg := NewAggregator(fixedNow())
	result := agg.Aggregate([]ChangeEvent{})
	if !result.ScannedAt.Equal(expected) {
		t.Errorf("expected ScannedAt %v, got %v", expected, result.ScannedAt)
	}
}

func TestAggregator_DefaultNow(t *testing.T) {
	before := time.Now()
	agg := NewAggregator(nil)
	result := agg.Aggregate(nil)
	after := time.Now()
	if result.ScannedAt.Before(before) || result.ScannedAt.After(after) {
		t.Errorf("ScannedAt %v not within expected range [%v, %v]", result.ScannedAt, before, after)
	}
}
