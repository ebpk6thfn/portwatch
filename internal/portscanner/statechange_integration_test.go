package portscanner

import (
	"testing"
	"time"
)

// TestStateChangeTracker_PipelineStyleUsage simulates how the tracker would be
// used in a pipeline: record on open, forget on close.
func TestStateChangeTracker_PipelineStyleUsage(t *testing.T) {
	base := time.Unix(5000, 0)
	current := base
	nowFn := func() time.Time { return current }
	tr := NewStateChangeTracker(nowFn)

	// Simulate port opened
	events := []ChangeEvent{
		{Entry: Entry{Port: 80, Proto: "tcp"}, Type: EventOpened},
		{Entry: Entry{Port: 443, Proto: "tcp"}, Type: EventOpened},
	}

	for _, e := range events {
		tr.Record(e.Entry.Key())
	}

	current = base.Add(30 * time.Second)

	// Port 80 closes
	close80 := ChangeEvent{Entry: Entry{Port: 80, Proto: "tcp"}, Type: EventClosed}
	_, dur := tr.Record(close80.Entry.Key())
	if dur < 30*time.Second {
		t.Errorf("expected duration >= 30s for port 80, got %v", dur)
	}
	tr.Forget(close80.Entry.Key())

	if tr.Len() != 1 {
		t.Errorf("expected 1 tracked key after close, got %d", tr.Len())
	}

	_, ok := tr.FirstSeen(Entry{Port: 443, Proto: "tcp"}.Key())
	if !ok {
		t.Error("expected port 443 still tracked")
	}
}
