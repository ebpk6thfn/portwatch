package portscanner

import "strings"

// Severity represents the importance level of a port change event.
type Severity int

const (
	// SeverityLow indicates a non-critical port event (e.g. ephemeral ports).
	SeverityLow Severity = iota
	// SeverityMedium indicates a notable port event.
	SeverityMedium
	// SeverityHigh indicates a critical port event (e.g. well-known service ports).
	SeverityHigh
)

// String returns a human-readable label for the severity.
func (s Severity) String() string {
	switch s {
	case SeverityHigh:
		return "high"
	case SeverityMedium:
		return "medium"
	default:
		return "low"
	}
}

// wellKnownPorts are ports that warrant high-severity alerts.
var wellKnownPorts = map[uint16]bool{
	22: true, 80: true, 443: true, 3306: true,
	5432: true, 6379: true, 27017: true, 8080: true,
}

// Classifier assigns a Severity to a ChangeEvent.
type Classifier struct {
	highPorts map[uint16]bool
}

// NewClassifier returns a Classifier seeded with well-known ports.
// Additional high-severity ports may be supplied via extra.
func NewClassifier(extra []uint16) *Classifier {
	hp := make(map[uint16]bool, len(wellKnownPorts)+len(extra))
	for p := range wellKnownPorts {
		hp[p] = true
	}
	for _, p := range extra {
		hp[p] = true
	}
	return &Classifier{highPorts: hp}
}

// Classify returns the Severity for a given ChangeEvent.
func (c *Classifier) Classify(ev ChangeEvent) Severity {
	if c.highPorts[ev.Entry.Port] {
		return SeverityHigh
	}
	// Ports below 1024 that are not explicitly listed are medium.
	if ev.Entry.Port < 1024 {
		return SeverityMedium
	}
	// Treat any event whose process name looks like a server as medium.
	name := strings.ToLower(ev.Entry.Process)
	if name != "" && name != "unknown" {
		return SeverityMedium
	}
	return SeverityLow
}
