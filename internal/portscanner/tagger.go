package portscanner

// Tag represents a short descriptive label attached to a ChangeEvent.
type Tag string

const (
	TagWellKnown  Tag = "well-known"
	TagPrivileged Tag = "privileged"
	TagEphemeral  Tag = "ephemeral"
	TagUserDefined Tag = "user-defined"
)

// Tagger assigns one or more tags to a ChangeEvent based on port ranges
// and an optional user-supplied extra set.
type Tagger struct {
	extraPorts map[uint16]struct{}
}

// NewTagger creates a Tagger. extraPorts are user-configured ports that
// should receive the TagUserDefined tag.
func NewTagger(extraPorts []uint16) *Tagger {
	m := make(map[uint16]struct{}, len(extraPorts))
	for _, p := range extraPorts {
		m[p] = struct{}{}
	}
	return &Tagger{extraPorts: m}
}

// Tag returns the set of tags applicable to the event.
func (t *Tagger) Tag(event ChangeEvent) []Tag {
	port := event.Entry.Port
	var tags []Tag

	if _, ok := t.extraPorts[port]; ok {
		tags = append(tags, TagUserDefined)
	}

	switch {
	case port < 1024:
		tags = append(tags, TagPrivileged)
		if port <= 1023 && isWellKnownPort(port) {
			tags = append(tags, TagWellKnown)
		}
	case port >= 49152:
		tags = append(tags, TagEphemeral)
	}

	return tags
}

// TagAll annotates a slice of events, returning a map from event index to tags.
func (t *Tagger) TagAll(events []ChangeEvent) map[int][]Tag {
	out := make(map[int][]Tag, len(events))
	for i, e := range events {
		if tags := t.Tag(e); len(tags) > 0 {
			out[i] = tags
		}
	}
	return out
}

// isWellKnownPort returns true for a small set of common ports.
func isWellKnownPort(port uint16) bool {
	switch port {
	case 22, 25, 53, 80, 110, 143, 443, 465, 587, 993, 995:
		return true
	}
	return false
}
