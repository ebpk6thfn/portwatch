package portscanner

import "strings"

// Filter defines criteria for including or excluding port entries
// from scan results.
type Filter struct {
	// ExcludePorts is a set of local ports to ignore.
	ExcludePorts map[uint16]struct{}
	// Protocols limits results to the given protocols ("tcp", "udp").
	// An empty slice means all protocols are included.
	Protocols []string
	// ExcludeLoopback skips entries whose local address is 127.0.0.1 or ::1.
	ExcludeLoopback bool
}

// NewFilter constructs a Filter from the provided option functions.
func NewFilter(opts ...FilterOption) *Filter {
	f := &Filter{
		ExcludePorts: make(map[uint16]struct{}),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// FilterOption is a functional option for Filter.
type FilterOption func(*Filter)

// WithExcludePorts adds ports that should be silently ignored.
func WithExcludePorts(ports ...uint16) FilterOption {
	return func(f *Filter) {
		for _, p := range ports {
			f.ExcludePorts[p] = struct{}{}
		}
	}
}

// WithProtocols restricts scanning to the named protocols.
func WithProtocols(protos ...string) FilterOption {
	return func(f *Filter) {
		f.Protocols = protos
	}
}

// WithExcludeLoopback configures the filter to drop loopback addresses.
func WithExcludeLoopback(exclude bool) FilterOption {
	return func(f *Filter) {
		f.ExcludeLoopback = exclude
	}
}

// Apply returns only the entries that pass all filter criteria.
func (f *Filter) Apply(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if _, excluded := f.ExcludePorts[e.LocalPort]; excluded {
			continue
		}
		if f.ExcludeLoopback {
			ip := e.LocalAddr
			if ip == "127.0.0.1" || ip == "::1" {
				continue
			}
		}
		if len(f.Protocols) > 0 && !containsProto(f.Protocols, e.Protocol) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func containsProto(protos []string, proto string) bool {
	for _, p := range protos {
		if strings.EqualFold(p, proto) {
			return true
		}
	}
	return false
}
