package portscanner

import (
	"fmt"
	"net"
	"time"
)

// ProbeResult holds the outcome of a liveness probe on a port.
type ProbeResult struct {
	Entry     Entry
	Reachable bool
	Latency   time.Duration
	Error     string
}

// Prober attempts a TCP dial to verify a port is truly accepting connections.
type Prober struct {
	timeout time.Duration
}

// NewProber creates a Prober with the given dial timeout.
func NewProber(timeout time.Duration) *Prober {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Prober{timeout: timeout}
}

// Probe dials the entry's address and returns a ProbeResult.
func (p *Prober) Probe(e Entry) ProbeResult {
	addr := net.JoinHostPort(e.IP.String(), fmt.Sprintf("%d", e.Port))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, p.timeout)
	latency := time.Since(start)
	if err != nil {
		return ProbeResult{Entry: e, Reachable: false, Latency: latency, Error: err.Error()}
	}
	_ = conn.Close()
	return ProbeResult{Entry: e, Reachable: true, Latency: latency}
}

// ProbeAll probes every entry and returns results.
func (p *Prober) ProbeAll(entries []Entry) []ProbeResult {
	results := make([]ProbeResult, 0, len(entries))
	for _, e := range entries {
		results = append(results, p.Probe(e))
	}
	return results
}
