package daemon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/portscanner"
)

// stubNotifier counts how many times Notify is called.
type stubNotifier struct {
	calls atomic.Int32
}

func (s *stubNotifier) Notify(_ context.Context, _ portscanner.Event) error {
	s.calls.Add(1)
	return nil
}

func defaultTestConfig(interval time.Duration) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Interval = interval
	return cfg
}

func TestDaemon_RunCancelImmediately(t *testing.T) {
	cfg := defaultTestConfig(50 * time.Millisecond)
	n := &stubNotifier{}

	d, err := New(cfg, n)
	if err != nil {
		t.Skipf("skipping: cannot create scanner (may need /proc): %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	if err := d.Run(ctx); err != nil {
		t.Fatalf("expected nil error on clean shutdown, got: %v", err)
	}
}

func TestDaemon_RunTicksAndShutdown(t *testing.T) {
	cfg := defaultTestConfig(20 * time.Millisecond)
	n := &stubNotifier{}

	d, err := New(cfg, n)
	if err != nil {
		t.Skipf("skipping: cannot create scanner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()

	if err := d.Run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At least one tick should have fired within 75ms with a 20ms interval.
	// We can't assert notify calls without mocking the scanner, but we verify
	// the daemon exits cleanly without panicking.
}
