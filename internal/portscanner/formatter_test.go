package portscanner

import (
	"strings"
	"testing"
	"time"
)

func makeFormatterEvent(kind ChangeKind, ip string, port uint16, proto, proc string, pid int) ChangeEvent {
	return ChangeEvent{
		Kind: kind,
		Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Entry: Entry{
			IP:       ip,
			Port:     port,
			Protocol: proto,
			Process:  proc,
			PID:      pid,
		},
	}
}

func TestFormatter_ShortFormat_WithProcess(t *testing.T) {
	f := NewFormatter(FormatShort, time.UTC)
	e := makeFormatterEvent(PortOpened, "0.0.0.0", 8080, "tcp", "nginx", 42)
	out := f.Format(e)

	if !strings.Contains(out, "opened") {
		t.Errorf("expected 'opened' in short output, got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in short output, got: %s", out)
	}
	if !strings.Contains(out, "nginx") {
		t.Errorf("expected process name in short output, got: %s", out)
	}
}

func TestFormatter_ShortFormat_UnknownProcess(t *testing.T) {
	f := NewFormatter(FormatShort, time.UTC)
	e := makeFormatterEvent(PortClosed, "127.0.0.1", 9090, "tcp", "", 0)
	out := f.Format(e)

	if !strings.Contains(out, "unknown") {
		t.Errorf("expected 'unknown' for missing process, got: %s", out)
	}
}

func TestFormatter_LongFormat_ContainsFields(t *testing.T) {
	f := NewFormatter(FormatLong, time.UTC)
	e := makeFormatterEvent(PortOpened, "0.0.0.0", 443, "tcp", "caddy", 101)
	out := f.Format(e)

	for _, want := range []string{"Protocol:", "Address:", "Process:", "PID:", "Event:", "Time:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected field %q in long output, got:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "caddy") {
		t.Errorf("expected process name 'caddy' in long output")
	}
}

func TestFormatter_FormatAll_MultipleEvents(t *testing.T) {
	f := NewFormatter(FormatShort, time.UTC)
	events := []ChangeEvent{
		makeFormatterEvent(PortOpened, "0.0.0.0", 80, "tcp", "nginx", 1),
		makeFormatterEvent(PortClosed, "0.0.0.0", 3000, "tcp", "node", 2),
	}
	out := f.FormatAll(events)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines for 2 events, got %d: %q", len(lines), out)
	}
}

func TestFormatter_FormatAll_Empty(t *testing.T) {
	f := NewFormatter(FormatShort, time.UTC)
	out := f.FormatAll(nil)
	if out != "" {
		t.Errorf("expected empty string for no events, got %q", out)
	}
}

func TestNewFormatter_NilLocation_UsesLocal(t *testing.T) {
	f := NewFormatter(FormatShort, nil)
	if f.timeZone == nil {
		t.Error("expected non-nil timeZone when nil passed to NewFormatter")
	}
}
