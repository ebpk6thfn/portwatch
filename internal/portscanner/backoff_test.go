package portscanner

import (
	"testing"
	"time"
)

func TestBackoff_ExponentialDelay(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffExponential,
		BaseDelay: 1 * time.Second,
		MaxDelay:  60 * time.Second,
		Multipler: 2.0,
	}
	b := NewBackoff(p)

	d1 := b.Record("key")
	if d1 != 1*time.Second {
		t.Fatalf("expected 1s, got %v", d1)
	}

	d2 := b.Record("key")
	if d2 != 2*time.Second {
		t.Fatalf("expected 2s, got %v", d2)
	}

	d3 := b.Record("key")
	if d3 != 4*time.Second {
		t.Fatalf("expected 4s, got %v", d3)
	}
}

func TestBackoff_LinearDelay(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffLinear,
		BaseDelay: 2 * time.Second,
		MaxDelay:  60 * time.Second,
	}
	b := NewBackoff(p)

	d1 := b.Record("k")
	if d1 != 2*time.Second {
		t.Fatalf("expected 2s, got %v", d1)
	}
	d2 := b.Record("k")
	if d2 != 4*time.Second {
		t.Fatalf("expected 4s, got %v", d2)
	}
}

func TestBackoff_MaxDelayCapped(t *testing.T) {
	p := BackoffPolicy{
		Strategy:  BackoffExponential,
		BaseDelay: 1 * time.Second,
		MaxDelay:  5 * time.Second,
		Multipler: 10.0,
	}
	b := NewBackoff(p)

	b.Record("x")
	b.Record("x")
	d := b.Record("x")
	if d > 5*time.Second {
		t.Fatalf("delay %v exceeds max", d)
	}
}

func TestBackoff_ResetClearsAttempts(t *testing.T) {
	p := DefaultBackoffPolicy()
	b := NewBackoff(p)

	b.Record("svc")
	b.Record("svc")
	if b.Attempts("svc") != 2 {
		t.Fatal("expected 2 attempts")
	}

	b.Reset("svc")
	if b.Attempts("svc") != 0 {
		t.Fatal("expected 0 attempts after reset")
	}

	d := b.Record("svc")
	if d != 1*time.Second {
		t.Fatalf("expected base delay after reset, got %v", d)
	}
}

func TestBackoff_IndependentKeys(t *testing.T) {
	b := NewBackoff(DefaultBackoffPolicy())

	b.Record("a")
	b.Record("a")
	b.Record("b")

	if b.Attempts("a") != 2 {
		t.Fatal("expected 2 for a")
	}
	if b.Attempts("b") != 1 {
		t.Fatal("expected 1 for b")
	}
}
