package portscanner

import (
	"testing"
)

func makeFilterEntry(proto, addr string, port uint16) Entry {
	return Entry{Protocol: proto, Address: addr, Port: port}
}

func TestFilter_NoOptions_PassesAll(t *testing.T) {
	f := NewFilter()
	entries := []Entry{
		makeFilterEntry("tcp", "0.0.0.0", 80),
		makeFilterEntry("udp", "127.0.0.1", 53),
		makeFilterEntry("tcp", "192.168.1.1", 443),
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
	if got[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", got[0].Port)
	}
}

func TestFilter_ExcludeLoopback(t *testing.T) {
	f := NewFilter(WithExcludeLoopback())
	entries := []Entry{
		makeFilterEntry("tcp", "127.0.0.1", 3306),
		makeFilterEntry("tcp", "0.0.0.0", 3306),
		makeFilterEntry("tcp", "::1", 5432),
	}
	got := f.Apply(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Address != "0.0.0.0" {
		t.Errorf("unexpected address %s", got[0].Address)
	}
}

func TestFilter_Protocols(t *testing.T) {
	f := NewFilter(WithProtocols("tcp"))
	entries := []Entry{
		makeFilterEntry("tcp", "0.0.0.0", 80),
		makeFilterEntry("udp", "0.0.0.0", 53),
		makeFilterEntry("udp", "0.0.0.0", 123),
	}
	got := f.Apply(entries)
	if len(got) != 1 || got[0].Protocol != "tcp" {
		t.Fatalf("expected 1 tcp entry, got %d", len(got))
	}
}

func TestFilter_ExcludePrivate(t *testing.T) {
	f := NewFilter(WithExcludePrivate())
	entries := []Entry{
		makeFilterEntry("tcp", "192.168.0.1", 22),
		makeFilterEntry("tcp", "10.0.0.5", 22),
		makeFilterEntry("tcp", "172.16.0.1", 22),
		makeFilterEntry("tcp", "8.8.8.8", 22),
	}
	got := f.Apply(entries)
	if len(got) != 1 || got[0].Address != "8.8.8.8" {
		t.Fatalf("expected only public address, got %+v", got)
	}
}

func TestFilter_CombinedOptions(t *testing.T) {
	f := NewFilter(
		WithProtocols("tcp"),
		WithExcludeLoopback(),
		WithExcludePorts(22),
	)
	entries := []Entry{
		makeFilterEntry("tcp", "127.0.0.1", 8080), // loopback — excluded
		makeFilterEntry("udp", "0.0.0.0", 53),     // wrong proto — excluded
		makeFilterEntry("tcp", "0.0.0.0", 22),     // excluded port
		makeFilterEntry("tcp", "0.0.0.0", 443),    // passes
	}
	got := f.Apply(entries)
	if len(got) != 1 || got[0].Port != 443 {
		t.Fatalf("expected only port 443, got %+v", got)
	}
}
