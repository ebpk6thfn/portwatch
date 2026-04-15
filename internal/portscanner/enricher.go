package portscanner

// Enricher attaches derived metadata (severity, labels) to ChangeEvents.
type Enricher struct {
	classifier *Classifier
	labels     map[uint16]string
}

// EnrichedEvent wraps a ChangeEvent with additional metadata.
type EnrichedEvent struct {
	ChangeEvent
	Severity Severity
	Label    string // optional human-readable label for the port
}

// NewEnricher creates an Enricher with the given classifier and port-label map.
// portLabels maps a port number to a descriptive label, e.g. 443 → "HTTPS".
func NewEnricher(c *Classifier, portLabels map[uint16]string) *Enricher {
	if portLabels == nil {
		portLabels = make(map[uint16]string)
	}
	return &Enricher{classifier: c, labels: portLabels}
}

// Enrich converts a slice of ChangeEvents into EnrichedEvents.
func (e *Enricher) Enrich(events []ChangeEvent) []EnrichedEvent {
	out := make([]EnrichedEvent, 0, len(events))
	for _, ev := range events {
		out = append(out, EnrichedEvent{
			ChangeEvent: ev,
			Severity:    e.classifier.Classify(ev),
			Label:       e.labelFor(ev.Entry.Port),
		})
	}
	return out
}

// FilterBySeverity returns only those EnrichedEvents at or above minSeverity.
func FilterBySeverity(events []EnrichedEvent, min Severity) []EnrichedEvent {
	out := events[:0:0]
	for _, ev := range events {
		if ev.Severity >= min {
			out = append(out, ev)
		}
	}
	return out
}

func (e *Enricher) labelFor(port uint16) string {
	if l, ok := e.labels[port]; ok {
		return l
	}
	return ""
}
