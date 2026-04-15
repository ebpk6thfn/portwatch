package portscanner

import (
	"net"
	"testing"
	"time"
)

func makeSnapshotEntry(proto, ip string, port uint16) Entry {
	return Entry{
		Protocol:  proto,
		LocalIP:   net.ParseIP(ip),
		LocalPort: port,
		State:     "LISTEN",
	}
}

func TestNewSnapshot_Fields(t *testing.T) {
	entries := []Entry{
		makeSnapshotEntry("tcp", "0.0.0.0", 80),
		makeSnapshotEntry("tcp", "0.0.0.0", 443),
	}
	before := time.Now()
	snap := NewSnapshot(entries, 5*time.Millisecond)
	after := time.Now()

	if snap.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", snap.Len())
	}
	if snap.ScanDuration != 5*time.Millisecond {
		t.Errorf("unexpected scan duration: %v", snap.ScanDuration)
	}
	if snap.CapturedAt.Before(before) || snap.CapturedAt.After(after) {
		t.Errorf("CapturedAt %v outside expected range", snap.CapturedAt)
	}
}

func TestSnapshot_ToMap_Keys(t *testing.T) {
	entries := []Entry{
		makeSnapshotEntry("tcp", "0.0.0.0", 80),
		makeSnapshotEntry("udp", "0.0.0.0", 53),
	}
	snap := NewSnapshot(entries, 0)
	m := snap.ToMap()

	if len(m) != 2 {
		t.Fatalf("expected 2 map entries, got %d", len(m))
	}
	for _, e := range entries {
		if _, ok := m[e.Key()]; !ok {
			t.Errorf("key %q not found in map", e.Key())
		}
	}
}

func TestSnapshot_ToMap_Empty(t *testing.T) {
	snap := NewSnapshot(nil, 0)
	m := snap.ToMap()
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestSnapshot_FilteredEntries(t *testing.T) {
	entries := []Entry{
		makeSnapshotEntry("tcp", "127.0.0.1", 8080),
		makeSnapshotEntry("tcp", "0.0.0.0", 443),
		makeSnapshotEntry("udp", "0.0.0.0", 53),
	}
	snap := NewSnapshot(entries, 0)

	f := NewFilter(WithExcludeLoopback(), WithProtocols("tcp"))
	result := snap.FilteredEntries(f)

	if len(result) != 1 {
		t.Fatalf("expected 1 filtered entry, got %d", len(result))
	}
	if result[0].LocalPort != 443 {
		t.Errorf("expected port 443, got %d", result[0].LocalPort)
	}
}
