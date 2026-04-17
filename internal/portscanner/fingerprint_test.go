package portscanner

import (
	"testing"
)

func makeFPEntry(proto string, port uint16) Entry {
	return Entry{Protocol: proto, Port: port}
}

func TestFingerprint_SameEntries_SameHash(t *testing.T) {
	fb := NewFingerprintBuilder()
	entries := []Entry{makeFPEntry("tcp", 80), makeFPEntry("tcp", 443)}
	f1 := fb.Build(entries)
	f2 := fb.Build(entries)
	if f1.Hash != f2.Hash {
		t.Errorf("expected same hash, got %s vs %s", f1.Hash, f2.Hash)
	}
}

func TestFingerprint_OrderIndependent(t *testing.T) {
	fb := NewFingerprintBuilder()
	a := fb.Build([]Entry{makeFPEntry("tcp", 80), makeFPEntry("tcp", 443)})
	b := fb.Build([]Entry{makeFPEntry("tcp", 443), makeFPEntry("tcp", 80)})
	if a.Hash != b.Hash {
		t.Errorf("expected order-independent hash, got %s vs %s", a.Hash, b.Hash)
	}
}

func TestFingerprint_DifferentEntries_DifferentHash(t *testing.T) {
	fb := NewFingerprintBuilder()
	a := fb.Build([]Entry{makeFPEntry("tcp", 80)})
	b := fb.Build([]Entry{makeFPEntry("tcp", 8080)})
	if a.Hash == b.Hash {
		t.Error("expected different hashes for different entries")
	}
}

func TestFingerprint_Diff_Added(t *testing.T) {
	fb := NewFingerprintBuilder()
	old := fb.Build([]Entry{makeFPEntry("tcp", 80)})
	new_ := fb.Build([]Entry{makeFPEntry("tcp", 80), makeFPEntry("tcp", 443)})
	added, removed := old.Diff(new_)
	if len(added) != 1 || added[0] != "tcp:443" {
		t.Errorf("unexpected added: %v", added)
	}
	if len(removed) != 0 {
		t.Errorf("unexpected removed: %v", removed)
	}
}

func TestFingerprint_Diff_Removed(t *testing.T) {
	fb := NewFingerprintBuilder()
	old := fb.Build([]Entry{makeFPEntry("tcp", 80), makeFPEntry("tcp", 443)})
	new_ := fb.Build([]Entry{makeFPEntry("tcp", 80)})
	added, removed := old.Diff(new_)
	if len(added) != 0 {
		t.Errorf("unexpected added: %v", added)
	}
	if len(removed) != 1 || removed[0] != "tcp:443" {
		t.Errorf("unexpected removed: %v", removed)
	}
}

func TestFingerprint_EmptyEntries(t *testing.T) {
	fb := NewFingerprintBuilder()
	f := fb.Build(nil)
	if f.Hash == "" {
		t.Error("expected non-empty hash for empty entries")
	}
	if len(f.PortSet) != 0 {
		t.Errorf("expected empty port set, got %v", f.PortSet)
	}
}
