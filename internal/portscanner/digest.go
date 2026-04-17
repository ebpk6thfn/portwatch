package portscanner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// Digest produces a stable hash of a snapshot's entries, useful for
// quickly detecting whether the port state has changed between scans.
type Digest struct {
	hash string
	count int
}

// NewDigest computes a SHA-256 digest over the sorted keys of the provided entries.
func NewDigest(entries []Entry) Digest {
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, e.Key())
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintln(h, k)
	}

	return Digest{
		hash:  hex.EncodeToString(h.Sum(nil)),
		count: len(entries),
	}
}

// Hash returns the hex-encoded SHA-256 hash string.
func (d Digest) Hash() string { return d.hash }

// Count returns the number of entries that were hashed.
func (d Digest) Count() int { return d.count }

// Equal reports whether two digests represent identical port state.
func (d Digest) Equal(other Digest) bool { return d.hash == other.hash }

// String returns a short human-readable representation.
func (d Digest) String() string {
	short := d.hash
	if len(short) > 12 {
		short = short[:12]
	}
	return fmt.Sprintf("digest(%s… n=%d)", short, d.count)
}
