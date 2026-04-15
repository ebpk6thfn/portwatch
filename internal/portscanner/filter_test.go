package portscanner

import (
	"testing"
)

func makeFilterEntry(proto, addr string, port uint16) Entry {
	return Entry{
		Protocol:  proto,
		LocalAddr: addr,
		LocalPort: port,
	}
}

func TestFilter_NoOptions_PassesAll(t *testing.T) {
	f := NewFilter()
	entries := []Entry{
		makeFilterEntry("tcp", "0.0.0.0", 80),
		makeFilterEntry("udp", "127.0.0.1", 53),
	}
	got := f.Apply(entries)
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
}

func TestFilter_ExcludePorts(t *testing.T) {
	f := NewFilter(WithExcludePorts(80, 443))
	entries := []Entry{
		makeFilterEntry("tcp", "0.0.0.0", 80),
		makeFilterEntry("tcp", "0.0.0.0", 443),
		makeFilterEntry("tcp", "0.0.0.0", 8080),
	}
	got := f.Apply(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].LocalPort != 8080 {
		t.Errorf("expected port 8080, got %d", got[0].LocalPort)
	}
}

func TestFilter_ExcludeLoopback(t *testing.T) {
	f := NewFilter(WithExcludeLoopback(true))
	entries := []Entry{
		makeFilterEntry("tcp", "127.0.0.1", 9000),
		makeFilterEntry("tcp", "::1", 9001),
		makeFilterEntry("tcp", "0.0.0.0", 9002),
	}
	got := f.Apply(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].LocalPort != 9002 {
		t.Errorf("expected port 9002, got %d", got[0].LocalPort)
	}
}

func TestFilter_Protocols(t *testing.T) {
	f := NewFilter(WithProtocols("tcp"))
	entries := []Entry{
		makeFilterEntry("tcp", "0.0.0.0", 80),
		makeFilterEntry("udp", "0.0.0.0", 53),
		makeFilterEntry("TCP", "0.0.0.0", 443),
	}
	got := f.Apply(entries)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestFilter_CombinedOptions(t *testing.T) {
	f := NewFilter(
		WithExcludePorts(22),
		WithExcludeLoopback(true),
		WithProtocols("tcp"),
	)
	entries := []Entry{
		makeFilterEntry("tcp", "0.0.0.0", 22),    // excluded port
		makeFilterEntry("tcp", "127.0.0.1", 8080), // loopback
		makeFilterEntry("udp", "0.0.0.0", 53),     // wrong proto
		makeFilterEntry("tcp", "0.0.0.0", 3000),   // should pass
	}
	got := f.Apply(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].LocalPort != 3000 {
		t.Errorf("expected port 3000, got %d", got[0].LocalPort)
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	f := NewFilter(WithExcludePorts(80))
	got := f.Apply([]Entry{})
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}
