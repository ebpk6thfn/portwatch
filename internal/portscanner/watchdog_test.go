package portscanner

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchdog_BeatResetsStale(t *testing.T) {
	policy := WatchdogPolicy{
		MaxSilence:    100 * time.Millisecond,
		CheckInterval: 10 * time.Millisecond,
	}
	var fired int32
	wd := NewWatchdog(policy, func(_ time.Duration) {
		atomic.AddInt32(&fired, 1)
	})

	wd.Beat()
	if d := wd.StaleDuration(); d > 50*time.Millisecond {
		t.Fatalf("expected small stale duration after beat, got %v", d)
	}
}

func TestWatchdog_AlertFiredWhenStale(t *testing.T) {
	base := time.Now()
	offset := 0 * time.Second
	nowFn := func() time.Time { return base.Add(offset) }

	policy := WatchdogPolicy{
		MaxSilence:    30 * time.Second,
		CheckInterval: 5 * time.Millisecond,
	}
	var fired int32
	wd := NewWatchdog(policy, func(_ time.Duration) {
		atomic.AddInt32(&fired, 1)
	})
	wd.now = nowFn

	wd.Start()
	defer wd.Stop()

	// Advance clock past MaxSilence
	offset = 60 * time.Second
	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&fired) == 0 {
		t.Fatal("expected watchdog alert to fire when stale")
	}
}

func TestWatchdog_NoAlertWhenBeating(t *testing.T) {
	policy := WatchdogPolicy{
		MaxSilence:    50 * time.Millisecond,
		CheckInterval: 5 * time.Millisecond,
	}
	var fired int32
	wd := NewWatchdog(policy, func(_ time.Duration) {
		atomic.AddInt32(&fired, 1)
	})
	wd.Start()
	defer wd.Stop()

	// Keep beating every 10ms for 60ms total; MaxSilence is 50ms
	deadline := time.Now().Add(60 * time.Millisecond)
	for time.Now().Before(deadline) {
		wd.Beat()
		time.Sleep(10 * time.Millisecond)
	}

	if atomic.LoadInt32(&fired) != 0 {
		t.Fatal("expected no watchdog alert while beating regularly")
	}
}

func TestWatchdog_DefaultPolicy(t *testing.T) {
	p := DefaultWatchdogPolicy()
	if p.MaxSilence <= 0 {
		t.Fatal("expected positive MaxSilence")
	}
	if p.CheckInterval <= 0 {
		t.Fatal("expected positive CheckInterval")
	}
	if p.CheckInterval >= p.MaxSilence {
		t.Fatal("CheckInterval should be less than MaxSilence")
	}
}

func TestWatchdog_StaleDuration_GrowsOverTime(t *testing.T) {
	base := time.Now()
	offset := 0 * time.Second
	nowFn := func() time.Time { return base.Add(offset) }

	wd := NewWatchdog(DefaultWatchdogPolicy(), func(_ time.Duration) {})
	wd.now = nowFn
	wd.Beat()

	offset = 5 * time.Second
	d := wd.StaleDuration()
	if d < 4*time.Second || d > 6*time.Second {
		t.Fatalf("expected ~5s stale duration, got %v", d)
	}
}
