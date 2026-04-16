package portscanner

import (
	"testing"
	"time"
)

var baseTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func makeReplayEvent(offsetSec int, kind string, port uint16) ReplayEvent {
	return ReplayEvent{
		At: baseTime.Add(time.Duration(offsetSec) * time.Second),
		Event: ChangeEvent{
			Kind:  kind,
			Entry: Entry{Port: port, Proto: "tcp"},
		},
	}
}

func TestReplayer_All_ReturnsSorted(t *testing.T) {
	events := []ReplayEvent{
		makeReplayEvent(10, "opened", 8080),
		makeReplayEvent(0, "opened", 443),
		makeReplayEvent(5, "closed", 80),
	}
	r := NewReplayer(events)
	all := r.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].Event.Entry.Port != 443 || all[1].Event.Entry.Port != 80 || all[2].Event.Entry.Port != 8080 {
		t.Errorf("unexpected order: %v", all)
	}
}

func TestReplayer_Len(t *testing.T) {
	r := NewReplayer([]ReplayEvent{
		makeReplayEvent(0, "opened", 80),
		makeReplayEvent(1, "opened", 443),
	})
	if r.Len() != 2 {
		t.Errorf("expected 2, got %d", r.Len())
	}
}

func TestReplayer_Between(t *testing.T) {
	r := NewReplayer([]ReplayEvent{
		makeReplayEvent(0, "opened", 80),
		makeReplayEvent(5, "opened", 443),
		makeReplayEvent(10, "closed", 8080),
	})
	from := baseTime.Add(3 * time.Second)
	to := baseTime.Add(8 * time.Second)
	result := r.Between(from, to)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].Event.Entry.Port != 443 {
		t.Errorf("expected port 443, got %d", result[0].Event.Entry.Port)
	}
}

func TestReplayer_Since(t *testing.T) {
	r := NewReplayer([]ReplayEvent{
		makeReplayEvent(0, "opened", 80),
		makeReplayEvent(5, "opened", 443),
		makeReplayEvent(10, "closed", 8080),
	})
	result := r.Since(baseTime.Add(5 * time.Second))
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestReplayer_Empty(t *testing.T) {
	r := NewReplayer(nil)
	if r.Len() != 0 {
		t.Errorf("expected 0")
	}
	if len(r.Between(baseTime, baseTime.Add(time.Hour))) != 0 {
		t.Errorf("expected empty between")
	}
}
