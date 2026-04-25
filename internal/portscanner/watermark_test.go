package portscanner

import (
	"testing"
	"time"
)

func makeWatermarkNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func advanceWatermarkNow(fn *func() time.Time, d time.Duration) {
	old := (*fn)()
	*fn = func() time.Time { return old.Add(d) }
}

func TestWatermark_InitiallyNormal(t *testing.T) {
	now := makeWatermarkNow(time.Now())
	w := NewWatermark(DefaultWatermarkPolicy(), now)
	if w.State() != WatermarkNormal {
		t.Fatal("expected normal state initially")
	}
}

func TestWatermark_BreachesAtHighMark(t *testing.T) {
	p := WatermarkPolicy{HighMark: 5, LowMark: 2, Window: time.Minute, Cooldown: time.Minute}
	now := makeWatermarkNow(time.Now())
	w := NewWatermark(p, now)

	var state WatermarkState
	for i := 0; i < 5; i++ {
		state = w.Record()
	}
	if state != WatermarkBreached {
		t.Fatalf("expected breached after %d events, got %v", p.HighMark, state)
	}
}

func TestWatermark_NormalBelowHighMark(t *testing.T) {
	p := WatermarkPolicy{HighMark: 10, LowMark: 3, Window: time.Minute, Cooldown: time.Minute}
	now := makeWatermarkNow(time.Now())
	w := NewWatermark(p, now)

	for i := 0; i < 4; i++ {
		w.Record()
	}
	if w.State() != WatermarkNormal {
		t.Fatal("expected normal when below high mark")
	}
}

func TestWatermark_EvictsOldEvents(t *testing.T) {
	p := WatermarkPolicy{HighMark: 5, LowMark: 2, Window: 10 * time.Second, Cooldown: time.Minute}
	base := time.Now()
	current := base
	now := func() time.Time { return current }
	w := NewWatermark(p, now)

	// Fill to breach
	for i := 0; i < 5; i++ {
		w.Record()
	}
	if w.State() != WatermarkBreached {
		t.Fatal("expected breach")
	}

	// Advance past window so old events evict
	current = base.Add(15 * time.Second)
	if w.Depth() != 0 {
		t.Fatalf("expected 0 events after eviction, got %d", w.Depth())
	}
}

func TestWatermark_RecoveryAfterCooldown(t *testing.T) {
	p := WatermarkPolicy{HighMark: 3, LowMark: 1, Window: time.Minute, Cooldown: 30 * time.Second}
	base := time.Now()
	current := base
	now := func() time.Time { return current }
	w := NewWatermark(p, now)

	// Breach
	for i := 0; i < 3; i++ {
		w.Record()
	}
	if w.State() != WatermarkBreached {
		t.Fatal("expected breach")
	}

	// Advance so events evict below low mark, triggering cooldown start
	current = base.Add(61 * time.Second)
	w.Record()
	if w.State() != WatermarkBreached {
		t.Fatal("expected still breached during cooldown")
	}

	// Advance past cooldown
	current = current.Add(31 * time.Second)
	w.Record()
	if w.State() != WatermarkNormal {
		t.Fatal("expected normal after cooldown elapsed")
	}
}

func TestWatermark_DepthReflectsWindow(t *testing.T) {
	p := WatermarkPolicy{HighMark: 100, LowMark: 10, Window: 5 * time.Second, Cooldown: time.Minute}
	base := time.Now()
	current := base
	now := func() time.Time { return current }
	w := NewWatermark(p, now)

	w.Record()
	w.Record()
	if w.Depth() != 2 {
		t.Fatalf("expected depth 2, got %d", w.Depth())
	}

	current = base.Add(6 * time.Second)
	if w.Depth() != 0 {
		t.Fatalf("expected depth 0 after window, got %d", w.Depth())
	}
}
