package portscanner

import (
	"testing"
	"time"
)

func TestRateLimitMetrics_InitialZero(t *testing.T) {
	m := NewRateLimitMetrics()
	if m.Allowed() != 0 {
		t.Fatalf("expected 0 allowed, got %d", m.Allowed())
	}
	if m.Suppressed() != 0 {
		t.Fatalf("expected 0 suppressed, got %d", m.Suppressed())
	}
}

func TestRateLimitMetrics_RecordAllowed(t *testing.T) {
	m := NewRateLimitMetrics()
	m.RecordAllowed()
	m.RecordAllowed()
	if m.Allowed() != 2 {
		t.Fatalf("expected 2, got %d", m.Allowed())
	}
}

func TestRateLimitMetrics_RecordSuppressed(t *testing.T) {
	m := NewRateLimitMetrics()
	m.RecordSuppressed()
	if m.Suppressed() != 1 {
		t.Fatalf("expected 1, got %d", m.Suppressed())
	}
}

func TestRateLimitMetrics_Reset(t *testing.T) {
	m := NewRateLimitMetrics()
	m.RecordAllowed()
	m.RecordSuppressed()
	before := time.Now()
	m.Reset()
	if m.Allowed() != 0 || m.Suppressed() != 0 {
		t.Fatal("expected counters reset to zero")
	}
	if m.LastReset().Before(before) {
		t.Fatal("LastReset should be updated after Reset")
	}
}

func TestRateLimitMetrics_Summary(t *testing.T) {
	m := NewRateLimitMetrics()
	m.RecordAllowed()
	m.RecordAllowed()
	m.RecordSuppressed()
	s := m.Summary()
	if s.Allowed != 2 {
		t.Fatalf("expected 2 allowed in summary, got %d", s.Allowed)
	}
	if s.Suppressed != 1 {
		t.Fatalf("expected 1 suppressed in summary, got %d", s.Suppressed)
	}
	if s.LastReset.IsZero() {
		t.Fatal("LastReset should not be zero")
	}
}

func TestRateLimitMetrics_ConcurrentSafe(t *testing.T) {
	m := NewRateLimitMetrics()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.RecordAllowed()
				m.RecordSuppressed()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if m.Allowed() != 1000 {
		t.Fatalf("expected 1000 allowed, got %d", m.Allowed())
	}
}
