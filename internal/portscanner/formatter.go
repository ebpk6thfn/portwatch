package portscanner

import (
	"fmt"
	"strings"
	"time"
)

// FormatStyle controls how a ChangeEvent is rendered as a string.
type FormatStyle int

const (
	// FormatShort produces a compact single-line summary.
	FormatShort FormatStyle = iota
	// FormatLong produces a multi-line human-readable description.
	FormatLong
)

// Formatter converts ChangeEvents into human-readable strings.
type Formatter struct {
	style    FormatStyle
	timeZone *time.Location
}

// NewFormatter creates a Formatter with the given style.
// If loc is nil, time.Local is used.
func NewFormatter(style FormatStyle, loc *time.Location) *Formatter {
	if loc == nil {
		loc = time.Local
	}
	return &Formatter{style: style, timeZone: loc}
}

// Format renders a single ChangeEvent as a string.
func (f *Formatter) Format(e ChangeEvent) string {
	switch f.style {
	case FormatLong:
		return f.formatLong(e)
	default:
		return f.formatShort(e)
	}
}

// FormatAll renders a slice of ChangeEvents, one per line.
func (f *Formatter) FormatAll(events []ChangeEvent) string {
	lines := make([]string, 0, len(events))
	for _, e := range events {
		lines = append(lines, f.Format(e))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) formatShort(e ChangeEvent) string {
	proc := e.Entry.Process
	if proc == "" {
		proc = "unknown"
	}
	return fmt.Sprintf("[%s] %s %s:%d (%s)",
		e.Timestamp.In(f.timeZone).Format("15:04:05"),
		e.Kind,
		e.Entry.IP,
		e.Entry.Port,
		proc,
	)
}

func (f *Formatter) formatLong(e ChangeEvent) string {
	proc := e.Entry.Process
	if proc == "" {
		proc = "unknown"
	}
	return fmt.Sprintf(
		"Time:     %s\nEvent:    %s\nProtocol: %s\nAddress:  %s:%d\nProcess:  %s\nPID:      %d",
		e.Timestamp.In(f.timeZone).Format(time.RFC3339),
		e.Kind,
		e.Entry.Protocol,
		e.Entry.IP,
		e.Entry.Port,
		proc,
		e.Entry.PID,
	)
}
