package portscanner

import (
	"testing"
	"time"
)

func freshMetrics() *Metrics {
	m := &Metrics{}
	return m
}

func TestMetrics_InitialZero(t *testing.T) {
	m := freshMetrics()
	snap := m.Snapshot()
	if snap.ScansTotal != 0 || snap.EventsEmitted != 0 || snap.EventsDropped != 0 {
		t.Fatal("expected all counters to be zero initially")
	}
}

func TestMetrics_RecordScan(t *testing.T) {
	m := freshMetrics()
	at := time.Now()
	dur := 42 * time.Millisecond
	m.RecordScan(dur, at)

	snap := m.Snapshot()
	if snap.ScansTotal != 1 {
		t.Fatalf("expected ScansTotal=1, got %d", snap.ScansTotal)
	}
	if snap.LastScanDur != dur {
		t.Fatalf("expected LastScanDur=%v, got %v", dur, snap.LastScanDur)
	}
	if !snap.LastScanAt.Equal(at) {
		t.Fatalf("expected LastScanAt=%v, got %v", at, snap.LastScanAt)
	}
}

func TestMetrics_RecordEmitted(t *testing.T) {
	m := freshMetrics()
	m.RecordEmitted(3)
	m.RecordEmitted(2)
	snap := m.Snapshot()
	if snap.EventsEmitted != 5 {
		t.Fatalf("expected EventsEmitted=5, got %d", snap.EventsEmitted)
	}
}

func TestMetrics_RecordDropped(t *testing.T) {
	m := freshMetrics()
	m.RecordDropped(7)
	snap := m.Snapshot()
	if snap.EventsDropped != 7 {
		t.Fatalf("expected EventsDropped=7, got %d", snap.EventsDropped)
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := freshMetrics()
	m.RecordScan(10*time.Millisecond, time.Now())
	m.RecordEmitted(5)
	m.RecordDropped(3)
	m.Reset()

	snap := m.Snapshot()
	if snap.ScansTotal != 0 {
		t.Fatalf("expected ScansTotal=0 after reset, got %d", snap.ScansTotal)
	}
	if snap.EventsEmitted != 0 {
		t.Fatalf("expected EventsEmitted=0 after reset, got %d", snap.EventsEmitted)
	}
	if snap.EventsDropped != 0 {
		t.Fatalf("expected EventsDropped=0 after reset, got %d", snap.EventsDropped)
	}
	if !snap.LastScanAt.IsZero() {
		t.Fatal("expected LastScanAt to be zero after reset")
	}
}

func TestMetrics_GlobalSingleton(t *testing.T) {
	g1 := GlobalMetrics()
	g2 := GlobalMetrics()
	if g1 != g2 {
		t.Fatal("GlobalMetrics should return the same pointer")
	}
}
