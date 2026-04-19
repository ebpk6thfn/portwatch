package portscanner

import (
	"testing"
	"time"
)

// TestSilencer_PipelineStyleUsage simulates a pipeline where certain ports
// are silenced during a maintenance window and events are filtered accordingly.
func TestSilencer_PipelineStyleUsage(t *testing.T) {
	now := time.Now()
	s := NewSilencer(func() time.Time { return now })

	// Silence port 9200 (e.g. Elasticsearch) for 1 hour maintenance
	s.Silence(9200, now.Add(1*time.Hour))

	events := []ChangeEvent{
		{Entry: Entry{Port: 9200, Protocol: "tcp"}},
		{Entry: Entry{Port: 443, Protocol: "tcp"}},
		{Entry: Entry{Port: 80, Protocol: "tcp"}},
	}

	out := s.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events after filter, got %d", len(out))
	}
	for _, e := range out {
		if e.Entry.Port == 9200 {
			t.Fatal("silenced port 9200 should not appear in output")
		}
	}

	// Advance time past silence window
	now = now.Add(2 * time.Hour)
	out2 := s.Filter(events)
	if len(out2) != 3 {
		t.Fatalf("expected all 3 events after silence expires, got %d", len(out2))
	}
}
