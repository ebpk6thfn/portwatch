package portscanner

import (
	"testing"
	"time"
)

func makeDebounceNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestDebouncer_FirstEventAllowed(t *testing.T) {
	now := time.Now()
	d := NewDebouncer(2 * time.Second)
	d.now = makeDebounceNow(now)

	if !d.Allow("tcp:80") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestDebouncer_SecondEventWithinWindow_Suppressed(t *testing.T) {
	now := time.Now()
	d := NewDebouncer(2 * time.Second)
	d.now = makeDebounceNow(now)

	d.Allow("tcp:80")
	// same instant — within window
	if d.Allow("tcp:80") {
		t.Fatal("expected second event within window to be suppressed")
	}
}

func TestDebouncer_EventAfterWindow_Allowed(t *testing.T) {
	now := time.Now()
	d := NewDebouncer(2 * time.Second)
	d.now = makeDebounceNow(now)

	d.Allow("tcp:80")
	d.now = makeDebounceNow(now.Add(3 * time.Second))

	if !d.Allow("tcp:80") {
		t.Fatal("expected event after window to be allowed")
	}
}

func TestDebouncer_IndependentKeys(t *testing.T) {
	now := time.Now()
	d := NewDebouncer(2 * time.Second)
	d.now = makeDebounceNow(now)

	d.Allow("tcp:80")
	if !d.Allow("tcp:443") {
		t.Fatal("expected independent key to be allowed")
	}
}

func TestDebouncer_Flush_RemovesExpired(t *testing.T) {
	now := time.Now()
	d := NewDebouncer(1 * time.Second)
	d.now = makeDebounceNow(now)

	d.Allow("tcp:80")
	d.Allow("tcp:443")

	if d.Len() != 2 {
		t.Fatalf("expected 2 tracked keys, got %d", d.Len())
	}

	d.now = makeDebounceNow(now.Add(2 * time.Second))
	d.Flush()

	if d.Len() != 0 {
		t.Fatalf("expected 0 tracked keys after flush, got %d", d.Len())
	}
}

func TestDebouncer_Flush_KeepsActive(t *testing.T) {
	now := time.Now()
	d := NewDebouncer(5 * time.Second)
	d.now = makeDebounceNow(now)

	d.Allow("tcp:80")
	d.now = makeDebounceNow(now.Add(2 * time.Second))
	d.Flush()

	if d.Len() != 1 {
		t.Fatalf("expected 1 active key after flush, got %d", d.Len())
	}
}
