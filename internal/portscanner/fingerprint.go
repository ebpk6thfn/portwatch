package portscanner

import (
	"fmt"
	"sort"
	"strings"
)

// Fingerprint represents a stable identity for a set of open ports at a point in time.
type Fingerprint struct {
	Hash    string
	PortSet []string
}

// FingerprintBuilder builds fingerprints from snapshots for change correlation.
type FingerprintBuilder struct{}

// NewFingerprintBuilder returns a new FingerprintBuilder.
func NewFingerprintBuilder() *FingerprintBuilder {
	return &FingerprintBuilder{}
}

// Build creates a Fingerprint from a slice of entries.
func (fb *FingerprintBuilder) Build(entries []Entry) Fingerprint {
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, entryFingerprintKey(e))
	}
	sort.Strings(keys)
	hash := stableHash(keys)
	return Fingerprint{
		Hash:    hash,
		PortSet: keys,
	}
}

// Diff returns keys added and removed between two fingerprints.
func (fp Fingerprint) Diff(other Fingerprint) (added, removed []string) {
	curr := toSet(fp.PortSet)
	next := toSet(other.PortSet)
	for k := range next {
		if !curr[k] {
			added = append(added, k)
		}
	}
	for k := range curr {
		if !next[k] {
			removed = append(removed, k)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return
}

func entryFingerprintKey(e Entry) string {
	return fmt.Sprintf("%s:%d", e.Protocol, e.Port)
}

func stableHash(keys []string) string {
	return fmt.Sprintf("%x", fnv32(strings.Join(keys, "|")))
}

func fnv32(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func toSet(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}
