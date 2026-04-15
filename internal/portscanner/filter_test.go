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
		makeFilterEntry("tcp", "192.168.1.5", 443),
	}
	for _, e := range entries {
		if !f.Apply(e) {
			t.Errorf("expected entry %v to pass with no options", e)
		}
	}
}

func TestFilter_ExcludePorts(t *testing.T) {
	f := NewFilter(WithExcludePorts(80, 443))

	if f.Apply(makeFilterEntry("tcp", "0.0.0.0", 80)) {
		t.Error("port 80 should be excluded")
	}
	if f.Apply(makeFilterEntry("tcp", "0.0.0.0", 443)) {
		t.Error("port 443 should be excluded")
	}
	if !f.Apply(makeFilterEntry("tcp", "0.0.0.0", 8080)) {
		t.Error("port 8080 should pass")
	}
}

func TestFilter_ExcludeLoopback(t *testing.T) {
	f := NewFilter(WithExcludeLoopback())

	if f.Apply(makeFilterEntry("tcp", "127.0.0.1", 3306)) {
		t.Error("loopback address should be excluded")
	}
	if !f.Apply(makeFilterEntry("tcp", "0.0.0.0", 3306)) {
		t.Error("non-loopback address should pass")
	}
}

func TestFilter_Protocols(t *testing.T) {
	f := NewFilter(WithProtocols("tcp"))

	if !f.Apply(makeFilterEntry("tcp", "0.0.0.0", 80)) {
		t.Error("tcp entry should pass")
	}
	if f.Apply(makeFilterEntry("udp", "0.0.0.0", 53)) {
		t.Error("udp entry should be filtered out")
	}
}

func TestFilter_ExcludePrivate(t *testing.T) {
	f := NewFilter(WithExcludePrivate())

	privateAddrs := []string{"10.0.0.1", "172.16.5.10", "192.168.0.100"}
	for _, addr := range privateAddrs {
		if f.Apply(makeFilterEntry("tcp", addr, 8080)) {
			t.Errorf("private address %s should be excluded", addr)
		}
	}
	if !f.Apply(makeFilterEntry("tcp", "8.8.8.8", 8080)) {
		t.Error("public address should pass")
	}
}

func TestFilter_CombinedOptions(t *testing.T) {
	f := NewFilter(
		WithProtocols("tcp"),
		WithExcludePorts(22),
		WithExcludeLoopback(),
	)

	// udp filtered by protocol
	if f.Apply(makeFilterEntry("udp", "0.0.0.0", 80)) {
		t.Error("udp should be filtered")
	}
	// port 22 excluded
	if f.Apply(makeFilterEntry("tcp", "0.0.0.0", 22)) {
		t.Error("port 22 should be excluded")
	}
	// loopback excluded
	if f.Apply(makeFilterEntry("tcp", "127.0.0.1", 8080)) {
		t.Error("loopback should be excluded")
	}
	// valid entry
	if !f.Apply(makeFilterEntry("tcp", "0.0.0.0", 8080)) {
		t.Error("valid entry should pass all filters")
	}
}
