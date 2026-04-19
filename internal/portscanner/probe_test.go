package portscanner

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func TestProber_ReachablePort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		conn, _ := ln.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	e := Entry{IP: addr.IP, Port: uint16(addr.Port), Protocol: "tcp"}

	p := NewProber(time.Second)
	r := p.Probe(e)

	if !r.Reachable {
		t.Errorf("expected reachable, got error: %s", r.Error)
	}
	if r.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestProber_UnreachablePort(t *testing.T) {
	// Use a port that is almost certainly not listening.
	ip := net.ParseIP("127.0.0.1")
	e := Entry{IP: ip, Port: 1, Protocol: "tcp"}

	p := NewProber(200 * time.Millisecond)
	r := p.Probe(e)

	if r.Reachable {
		t.Error("expected unreachable")
	}
	if r.Error == "" {
		t.Error("expected non-empty error string")
	}
}

func TestProber_DefaultTimeout(t *testing.T) {
	p := NewProber(0)
	if p.timeout != 2*time.Second {
		t.Errorf("expected default 2s, got %v", p.timeout)
	}
}

func TestProber_ProbeAll(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	portStr := strconv.Itoa(addr.Port)
	_ = portStr

	e1 := Entry{IP: addr.IP, Port: uint16(addr.Port), Protocol: "tcp"}
	e2 := Entry{IP: net.ParseIP("127.0.0.1"), Port: 1, Protocol: "tcp"}

	p := NewProber(300 * time.Millisecond)
	results := p.ProbeAll([]Entry{e1, e2})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Reachable {
		t.Error("first entry should be reachable")
	}
	if results[1].Reachable {
		t.Error("second entry should not be reachable")
	}
}
