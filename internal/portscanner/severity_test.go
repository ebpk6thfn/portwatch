package portscanner

import (
	"testing"
)

func makeSeverityEvent(port uint16, process string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:    port,
			Proto:   "tcp",
			Address: "0.0.0.0",
			Process: process,
		},
		Kind: EventOpened,
	}
}

func TestClassifier_WellKnownPort_IsHigh(t *testing.T) {
	c := NewClassifier(nil)
	ev := makeSeverityEvent(443, "nginx")
	if got := c.Classify(ev); got != SeverityHigh {
		t.Errorf("expected high, got %s", got)
	}
}

func TestClassifier_ExtraPort_IsHigh(t *testing.T) {
	c := NewClassifier([]uint16{9200})
	ev := makeSeverityEvent(9200, "elasticsearch")
	if got := c.Classify(ev); got != SeverityHigh {
		t.Errorf("expected high, got %s", got)
	}
}

func TestClassifier_PrivilegedUnknownPort_IsMedium(t *testing.T) {
	c := NewClassifier(nil)
	ev := makeSeverityEvent(512, "unknown")
	// port < 1024, not in well-known list, process is "unknown" → medium via port rule
	if got := c.Classify(ev); got != SeverityMedium {
		t.Errorf("expected medium, got %s", got)
	}
}

func TestClassifier_NamedProcess_IsMedium(t *testing.T) {
	c := NewClassifier(nil)
	ev := makeSeverityEvent(8888, "myserver")
	if got := c.Classify(ev); got != SeverityMedium {
		t.Errorf("expected medium, got %s", got)
	}
}

func TestClassifier_EphemeralUnknown_IsLow(t *testing.T) {
	c := NewClassifier(nil)
	ev := makeSeverityEvent(54321, "")
	if got := c.Classify(ev); got != SeverityLow {
		t.Errorf("expected low, got %s", got)
	}
}

func TestSeverity_String(t *testing.T) {
	cases := []struct {
		s    Severity
		want string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
