package portscanner

import (
	"testing"
)

func makeDigestEntry(proto, ip string, port uint16) Entry {
	return Entry{Protocol: proto, IP: ip, Port: port}
}

func TestDigest_EmptyEntries(t *testing.T) {
	d := NewDigest(nil)
	if d.Count() != 0 {
		t.Fatalf("expected count 0, got %d", d.Count())
	}
	if d.Hash() == "" {
		t.Fatal("expected non-empty hash for empty input")
	}
}

func TestDigest_SameEntriesSameHash(t *testing.T) {
	entries := []Entry{
		makeDigestEntry("tcp", "0.0.0.0", 80),
		makeDigestEntry("tcp", "0.0.0.0", 443),
	}
	d1 := NewDigest(entries)
	d2 := NewDigest(entries)
	if !d1.Equal(d2) {
		t.Fatalf("expected equal digests, got %s vs %s", d1.Hash(), d2.Hash())
	}
}

func TestDigest_OrderIndependent(t *testing.T) {
	a := []Entry{makeDigestEntry("tcp", "0.0.0.0", 80), makeDigestEntry("udp", "0.0.0.0", 53)}
	b := []Entry{makeDigestEntry("udp", "0.0.0.0", 53), makeDigestEntry("tcp", "0.0.0.0", 80)}
	if !NewDigest(a).Equal(NewDigest(b)) {
		t.Fatal("digest should be order-independent")
	}
}

func TestDigest_DifferentEntriesDifferentHash(t *testing.T) {
	d1 := NewDigest([]Entry{makeDigestEntry("tcp", "0.0.0.0", 80)})
	d2 := NewDigest([]Entry{makeDigestEntry("tcp", "0.0.0.0", 8080)})
	if d1.Equal(d2) {
		t.Fatal("expected different digests for different entries")
	}
}

func TestDigest_CountMatchesInput(t *testing.T) {
	entries := []Entry{
		makeDigestEntry("tcp", "127.0.0.1", 22),
		makeDigestEntry("tcp", "127.0.0.1", 8080),
		makeDigestEntry("udp", "0.0.0.0", 123),
	}
	d := NewDigest(entries)
	if d.Count() != 3 {
		t.Fatalf("expected count 3, got %d", d.Count())
	}
}

func TestDigest_String_ContainsHash(t *testing.T) {
	d := NewDigest([]Entry{makeDigestEntry("tcp", "0.0.0.0", 80)})
	s := d.String()
	if len(s) == 0 {
		t.Fatal("String() should not be empty")
	}
	if s[:7] != "digest(" {
		t.Fatalf("unexpected String() format: %s", s)
	}
}
