package portscanner

import (
	"testing"
	"time"
)

// TestScoreboard_PipelineStyleScoring simulates events flowing through a
// pipeline where each opened port contributes to a scoreboard and the
// top offenders are surfaced at the end of the window.
func TestScoreboard_PipelineStyleScoring(t *testing.T) {
	sb := NewScoreboard(5 * time.Minute)

	events := []struct {
		key   string
		delta float64
	}{
		{"tcp:22", 1.0},
		{"tcp:80", 2.0},
		{"tcp:22", 1.0},
		{"udp:53", 0.5},
		{"tcp:80", 2.0},
		{"tcp:80", 2.0},
	}

	for _, ev := range events {
		sb.Add(ev.key, ev.delta)
	}

	top := sb.Top(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 top entries, got %d", len(top))
	}
	if top[0].Key != "tcp:80" {
		t.Errorf("expected tcp:80 at top, got %s", top[0].Key)
	}
	if top[0].Score != 6.0 {
		t.Errorf("expected score 6.0, got %.2f", top[0].Score)
	}
	if top[0].Hits != 3 {
		t.Errorf("expected 3 hits, got %d", top[0].Hits)
	}
	if top[1].Key != "tcp:22" {
		t.Errorf("expected tcp:22 second, got %s", top[1].Key)
	}
}
