package portscanner

import (
	"testing"
	"time"
)

func makeGraceNow(base time.Time) func() time.Time {
	current := base
	return func() time.Time { return current }
}

func TestGrace_InWindowSuppressesEvent(t *testing.T) {
	base := time.Now()
	nowFn := makeGraceNow(base)
	policy := GracePolicy{Window: 10 * time.Second}
	g := newGraceWithClock(policy, nowFn)

	if g.Allow(ChangeEvent{}) {
		t.Fatal("expected event to be suppressed during grace window")
	}
}

func TestGrace_AfterWindowAllowsEvent(t *testing.T) {
	base := time.Now()
	current := base
	nowFn := func() time.Time { return current }
	policy := GracePolicy{Window: 5 * time.Second}
	g := newGraceWithClock(policy, nowFn)

	current = base.Add(6 * time.Second)

	if !g.Allow(ChangeEvent{}) {
		t.Fatal("expected event to be allowed after grace window")
	}
}

func TestGrace_ExactBoundary_StillSuppressed(t *testing.T) {
	base := time.Now()
	current := base
	nowFn := func() time.Time { return current }
	policy := GracePolicy{Window: 5 * time.Second}
	g := newGraceWithClock(policy, nowFn)

	current = base.Add(4 * time.Second)

	if g.Allow(ChangeEvent{}) {
		t.Fatal("expected event to be suppressed just before window expires")
	}
}

func TestGrace_Filter_DropsEventsInWindow(t *testing.T) {
	base := time.Now()
	nowFn := makeGraceNow(base)
	policy := GracePolicy{Window: 10 * time.Second}
	g := newGraceWithClock(policy, nowFn)

	events := []ChangeEvent{{}, {}, {}}
	result := g.Filter(events)

	if len(result) != 0 {
		t.Fatalf("expected 0 events, got %d", len(result))
	}
}

func TestGrace_Filter_PassesEventsAfterWindow(t *testing.T) {
	base := time.Now()
	current := base.Add(20 * time.Second)
	nowFn := func() time.Time { return current }
	policy := GracePolicy{Window: 5 * time.Second}
	g := newGraceWithClock(policy, nowFn)

	events := []ChangeEvent{{}, {}}
	result := g.Filter(events)

	if len(result) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result))
	}
}

func TestGrace_InWindow_ReportsCorrectly(t *testing.T) {
	base := time.Now()
	current := base
	nowFn := func() time.Time { return current }
	policy := GracePolicy{Window: 5 * time.Second}
	g := newGraceWithClock(policy, nowFn)

	if !g.InWindow() {
		t.Fatal("expected InWindow to be true at startup")
	}

	current = base.Add(10 * time.Second)
	if g.InWindow() {
		t.Fatal("expected InWindow to be false after window expires")
	}
}

func TestGrace_DefaultPolicy_IsValid(t *testing.T) {
	p := DefaultGracePolicy()
	if p.Window <= 0 {
		t.Fatalf("expected positive window, got %v", p.Window)
	}
}
