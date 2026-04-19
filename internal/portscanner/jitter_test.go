package portscanner

import (
	"testing"
	"time"
)

func TestJitter_ZeroFactor_ReturnsSame(t *testing.T) {
	j := NewJitter(0)
	base := 100 * time.Millisecond
	for i := 0; i < 20; i++ {
		if got := j.Apply(base); got != base {
			t.Fatalf("expected %v, got %v", base, got)
		}
	}
}

func TestJitter_Apply_WithinBounds(t *testing.T) {
	j := NewJitter(0.2)
	base := 1 * time.Second
	low := time.Duration(float64(base) * 0.8)
	high := time.Duration(float64(base) * 1.2)
	for i := 0; i < 100; i++ {
		got := j.Apply(base)
		if got < low || got > high {
			t.Fatalf("Apply(%v) = %v, want in [%v, %v]", base, got, low, high)
		}
	}
}

func TestJitter_ApplyPositive_NeverLessThanBase(t *testing.T) {
	j := NewJitter(0.3)
	base := 500 * time.Millisecond
	for i := 0; i < 100; i++ {
		got := j.ApplyPositive(base)
		if got < base {
			t.Fatalf("ApplyPositive returned %v < base %v", got, base)
		}
	}
}

func TestJitter_ApplyPositive_WithinUpperBound(t *testing.T) {
	j := NewJitter(0.5)
	base := 200 * time.Millisecond
	high := time.Duration(float64(base) * 1.5)
	for i := 0; i < 100; i++ {
		got := j.ApplyPositive(base)
		if got > high {
			t.Fatalf("ApplyPositive(%v) = %v, exceeds upper bound %v", base, got, high)
		}
	}
}

func TestJitter_NegativeBase_ReturnsBase(t *testing.T) {
	j := NewJitter(0.2)
	base := -1 * time.Second
	if got := j.Apply(base); got != base {
		t.Fatalf("expected %v for negative base, got %v", base, got)
	}
}

func TestJitter_FactorClamped_Above1(t *testing.T) {
	j := NewJitter(5.0)
	if j.factor != 1.0 {
		t.Fatalf("expected factor clamped to 1.0, got %v", j.factor)
	}
}

func TestJitter_FactorClamped_BelowZero(t *testing.T) {
	j := NewJitter(-0.5)
	if j.factor != 0 {
		t.Fatalf("expected factor clamped to 0, got %v", j.factor)
	}
}
