package portscanner

import "net"

// FilterOption is a functional option for configuring a Filter.
type FilterOption func(*Filter)

// Filter decides which port entries are relevant for monitoring.
type Filter struct {
	excludePorts    map[uint16]bool
	protocols       map[string]bool
	excludeLoopback bool
	excludePrivate  bool
}

// NewFilter creates a Filter with the provided options applied.
func NewFilter(opts ...FilterOption) *Filter {
	f := &Filter{
		excludePorts: make(map[uint16]bool),
		protocols:   make(map[string]bool),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// WithExcludePorts excludes specific port numbers from results.
func WithExcludePorts(ports ...uint16) FilterOption {
	return func(f *Filter) {
		for _, p := range ports {
			f.excludePorts[p] = true
		}
	}
}

// WithProtocols restricts results to the given protocols (e.g. "tcp", "udp").
func WithProtocols(protos ...string) FilterOption {
	return func(f *Filter) {
		for _, p := range protos {
			f.protocols[p] = true
		}
	}
}

// WithExcludeLoopback drops entries whose address is a loopback address.
func WithExcludeLoopback() FilterOption {
	return func(f *Filter) { f.excludeLoopback = true }
}

// WithExcludePrivate drops entries whose address falls in an RFC-1918 range.
func WithExcludePrivate() FilterOption {
	return func(f *Filter) { f.excludePrivate = true }
}

// Apply returns true when the entry should be kept.
func (f *Filter) Apply(e Entry) bool {
	if f.excludePorts[e.Port] {
		return false
	}
	if len(f.protocols) > 0 && !f.protocols[e.Protocol] {
		return false
	}
	ip := net.ParseIP(e.Address)
	if f.excludeLoopback && ip != nil && ip.IsLoopback() {
		return false
	}
	if f.excludePrivate && ip != nil && isPrivate(ip) {
		return false
	}
	return true
}

var privateRanges = []net.IPNet{
	{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
	{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
	{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
}

func isPrivate(ip net.IP) bool {
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}
