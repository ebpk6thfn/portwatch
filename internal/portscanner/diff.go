package portscanner

// ChangeType indicates whether a port was opened or closed.
type ChangeType string

const (
	PortOpened ChangeType = "opened"
	PortClosed ChangeType = "closed"
)

// PortChange describes a single change in port state.
type PortChange struct {
	Type  ChangeType
	Entry PortEntry
}

// Diff compares two port snapshots and returns a list of changes.
// prev is the previous snapshot; curr is the current snapshot.
func Diff(prev, curr map[int]PortEntry) []PortChange {
	var changes []PortChange

	// Detect newly opened ports.
	for port, entry := range curr {
		if _, existed := prev[port]; !existed {
			changes = append(changes, PortChange{
				Type:  PortOpened,
				Entry: entry,
			})
		}
	}

	// Detect closed ports.
	for port, entry := range prev {
		if _, stillOpen := curr[port]; !stillOpen {
			changes = append(changes, PortChange{
				Type:  PortClosed,
				Entry: entry,
			})
		}
	}

	return changes
}
