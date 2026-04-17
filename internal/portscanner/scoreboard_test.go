package portscanner

import (
	"fmt"
	"testing"
	"time"
)

func makeScoreboardNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestScoreboard_AddAndTop(t *testing.T) {
	sb := NewScoreboard(time.Minute)
	sb.Add("a", 3.0)
	sb.Add("b", 1.0)
	sb.Add("a", 2.0)

	top := sb.Top(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(top))
	}
	if top[0].Key != "a" || top[0].Score != 5.0 {
		t.Errorf("expected a=5.0, got %s", top[0])
	}
	if top[1].Key != "b" || top[1].Score != 1.0 {
		t.Errorf("expected b=1.0, got %s", top[1])
	}
}

func TestScoreboard_HitsTracked(t *testing.T) {
	sb := NewScoreboard(time.Minute)
	sb.Add("x", 1.0)
	sb.Add("x", 1.0)
	sb.Add("x", 1.0)

	top := sb.Top(1)
	if top[0].Hits != 3 {
		t.Errorf("expected 3 hits, got %d", top[0].Hits)
	}
}

func TestScoreboard_EvictsExpired(t *testing.T) {
	base := time.Now()
	sb := NewScoreboard(time.Minute)
	nowFn := makeScoreboardNow(base)
	sb.now = nowFn

	sb.Add("old", 10.0)

	// advance time beyond TTL
	sb.now = func() time.Time { return base.Add(2 * time.Minute) }
	sb.Add("new", 1.0)

	if sb.Len() != 1 {
		t.Errorf("expected 1 entry after eviction, got %d", sb.Len())
	}
	top := sb.Top(5)
	if top[0].Key != "new" {
		t.Errorf("expected new entry, got %s", top[0].Key)
	}
}

func TestScoreboard_TopLimitedToN(t *testing.T) {
	sb := NewScoreboard(time.Minute)
	for i := 0; i < 10; i++ {
		sb.Add(fmt.Sprintf("key%d", i), float64(i))
	}
	top := sb.Top(3)
	if len(top) != 3 {
		t.Errorf("expected 3, got %d", len(top))
	}
}

func TestScoreboard_EmptyTop(t *testing.T) {
	sb := NewScoreboard(time.Minute)
	if got := sb.Top(5); len(got) != 0 {
		t.Errorf("expected empty top, got %d entries", len(got))
	}
}

func TestScoreEntry_String(t *testing.T) {
	e := ScoreEntry{Key: "tcp:80", Score: 4.5, Hits: 2}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
