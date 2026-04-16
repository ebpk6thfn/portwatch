package portscanner

import (
	"testing"
	"time"
)

func makeCounterNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestCounter_FirstAdd_ReturnsOne(t *testing.T) {
	c := NewCounter(time.Minute)
	n := c.Add("tcp:80")
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
}

func TestCounter_MultipleAdds_Accumulate(t *testing.T) {
	c := NewCounter(time.Minute)
	c.Add("tcp:80")
	c.Add("tcp:80")
	n := c.Add("tcp:80")
	if n != 3 {
		t.Fatalf("expected 3, got %d", n)
	}
}

func TestCounter_EvictsExpired(t *testing.T) {
	base := time.Now()
	c := NewCounter(time.Minute)

	// record at base
	c.now = func() time.Time { return base }
	c.Add("tcp:443")
	c.Add("tcp:443")

	// advance past window
	c.now = func() time.Time { return base.Add(2 * time.Minute) }
	n := c.Count("tcp:443")
	if n != 0 {
		t.Fatalf("expected 0 after eviction, got %d", n)
	}
}

func TestCounter_Count_DoesNotAdd(t *testing.T) {
	c := NewCounter(time.Minute)
	c.Count("tcp:22")
	n := c.Count("tcp:22")
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestCounter_Reset_ClearsKey(t *testing.T) {
	c := NewCounter(time.Minute)
	c.Add("tcp:8080")
	c.Add("tcp:8080")
	c.Reset("tcp:8080")
	if n := c.Count("tcp:8080"); n != 0 {
		t.Fatalf("expected 0 after reset, got %d", n)
	}
}

func TestCounter_IndependentKeys(t *testing.T) {
	c := NewCounter(time.Minute)
	c.Add("tcp:80")
	c.Add("tcp:80")
	c.Add("udp:53")
	if n := c.Count("tcp:80"); n != 2 {
		t.Fatalf("tcp:80 expected 2, got %d", n)
	}
	if n := c.Count("udp:53"); n != 1 {
		t.Fatalf("udp:53 expected 1, got %d", n)
	}
}
