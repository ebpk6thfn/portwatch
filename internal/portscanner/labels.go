package portscanner

// WellKnownLabels maps common port numbers to human-readable service names.
var WellKnownLabels = map[uint16]string{
	20:   "FTP Data",
	21:   "FTP Control",
	22:   "SSH",
	23:   "Telnet",
	25:   "SMTP",
	53:   "DNS",
	80:   "HTTP",
	110:  "POP3",
	143:  "IMAP",
	443:  "HTTPS",
	465:  "SMTPS",
	587:  "SMTP Submission",
	993:  "IMAPS",
	995:  "POP3S",
	3306: "MySQL",
	5432: "PostgreSQL",
	6379: "Redis",
	8080: "HTTP Alt",
	8443: "HTTPS Alt",
	27017: "MongoDB",
}

// Labeler assigns human-readable labels to ChangeEvents based on port number.
type Labeler struct {
	extra map[uint16]string
}

// NewLabeler returns a Labeler optionally extended with caller-supplied mappings.
// Caller-supplied entries take precedence over WellKnownLabels.
func NewLabeler(extra map[uint16]string) *Labeler {
	merged := make(map[uint16]string, len(WellKnownLabels)+len(extra))
	for k, v := range WellKnownLabels {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return &Labeler{extra: merged}
}

// Label returns the service label for the given port, or an empty string if
// the port is not recognised.
func (l *Labeler) Label(port uint16) string {
	return l.extra[port]
}

// Annotate sets the Label field on every event in the slice and returns it.
func (l *Labeler) Annotate(events []ChangeEvent) []ChangeEvent {
	for i := range events {
		events[i].Label = l.Label(events[i].Port)
	}
	return events
}
