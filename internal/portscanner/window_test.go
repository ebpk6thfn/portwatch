package portscanner

import (
	"testing"
	"time"
)

func makeWindowNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestWindowCounter_FirstAdd_ReturnsOne(t *testing.T) {
	wc := NewWindowCounter(10 * time.Second)
	if got := wc.Add("k"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestWindowCounter_MultipleAdds_Accumulate(t *testing.T) {
	wc := NewWindowCounter(10 * time.Second)
	wc.Add("k")
	wc.Add("k")
	if got := wc.Add("k"); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestWindowCounter_EvictsExpired(t *testing.T) {
	base := time.Now()
	wc := NewWindowCounter(5 * time.Second)
	wc.nowFn = func() time.Time { return base }
	wc.Add("k")
	wc.Add("k")
	// advance past window
	wc.nowFn = func() time.Time { return base.Add(6 * time.Second) }
	wc.Add("k")
	if got := wc.Count("k"); got != 1 {
		t.Fatalf("expected 1 after eviction, got %d", got)
	}
}

func TestWindowCounter_Count_DoesNotAdd(t *testing.T) {
	wc := NewWindowCounter(10 * time.Second)
	wc.Add("k")
	wc.Count("k")
	if got := wc.Count("k"); got != 1 {
		t.Fatalf("Count should not add, expected 1, got %d", got)
	}
}

func TestWindowCounter_Reset_ClearsKey(t *testing.T) {
	wc := NewWindowCounter(10 * time.Second)
	wc.Add("k")
	wc.Add("k")
	wc.Reset("k")
	if got := wc.Count("k"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestWindowCounter_IndependentKeys(t *testing.T) {
	wc := NewWindowCounter(10 * time.Second)
	wc.Add("a")
	wc.Add("a")
	wc.Add("b")
	if got := wc.Count("a"); got != 2 {
		t.Fatalf("expected 2 for a, got %d", got)
	}
	if got := wc.Count("b"); got != 1 {
		t.Fatalf("expected 1 for b, got %d", got)
	}
}
