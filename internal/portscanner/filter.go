package portscanner

import (
	"net"
)

// FilterOption configures a Filter.
type FilterOption func(*Filter)

// Filter decides which port entries should be included in scan results.
type Filter struct {
	excludePorts    map[uint16]bool
	protocols       map[string]bool
	excludeLoopback bool
	excludePrivate  bool
}

// NewFilter creates a Filter with the given options applied.
func NewFilter(opts ...FilterOption) *Filter {
	f := &Filter{
		excludePorts: make(map[uint16]bool),
		protocols:    make(map[string]bool),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// WithExcludePorts excludes the given port numbers from results.
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

// WithExcludePrivate drops entries whose address falls in a private IP range.
func WithExcludePrivate() FilterOption {
	return func(f *Filter) { f.excludePrivate = true }
}

// Allow returns true when entry passes all active filter rules.
func (f *Filter) Allow(e Entry) bool {
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

// Apply returns only the entries that pass the filter.
func (f *Filter) Apply(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if f.Allow(e) {
			out = append(out, e)
		}
	}
	return out
}

var privateRanges []*net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateRanges = append(privateRanges, block)
	}
}

func isPrivate(ip net.IP) bool {
	for _, block := range privateRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
