package portscanner

import (
	"testing"
)

func TestLabeler_WellKnownPort(t *testing.T) {
	l := NewLabeler(nil)
	if got := l.Label(22); got != "SSH" {
		t.Errorf("expected SSH, got %q", got)
	}
	if got := l.Label(443); got != "HTTPS" {
		t.Errorf("expected HTTPS, got %q", got)
	}
}

func TestLabeler_UnknownPort_ReturnsEmpty(t *testing.T) {
	l := NewLabeler(nil)
	if got := l.Label(9999); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestLabeler_ExtraOverridesWellKnown(t *testing.T) {
	extra := map[uint16]string{
		22:   "Custom SSH",
		9000: "My Service",
	}
	l := NewLabeler(extra)
	if got := l.Label(22); got != "Custom SSH" {
		t.Errorf("expected Custom SSH, got %q", got)
	}
	if got := l.Label(9000); got != "My Service" {
		t.Errorf("expected My Service, got %q", got)
	}
}

func TestLabeler_ExtraDoesNotPolluteSibling(t *testing.T) {
	l1 := NewLabeler(map[uint16]string{8080: "Override"})
	l2 := NewLabeler(nil)
	if got := l1.Label(8080); got != "Override" {
		t.Errorf("l1: expected Override, got %q", got)
	}
	if got := l2.Label(8080); got != "HTTP Alt" {
		t.Errorf("l2: expected HTTP Alt, got %q", got)
	}
}

func makeLabelEvent(port uint16, kind string) ChangeEvent {
	return ChangeEvent{Port: port, Kind: kind}
}

func TestLabeler_Annotate_SetsLabel(t *testing.T) {
	l := NewLabeler(nil)
	events := []ChangeEvent{
		makeLabelEvent(22, "opened"),
		makeLabelEvent(9999, "opened"),
		makeLabelEvent(3306, "closed"),
	}
	annotated := l.Annotate(events)
	if annotated[0].Label != "SSH" {
		t.Errorf("port 22: expected SSH, got %q", annotated[0].Label)
	}
	if annotated[1].Label != "" {
		t.Errorf("port 9999: expected empty, got %q", annotated[1].Label)
	}
	if annotated[2].Label != "MySQL" {
		t.Errorf("port 3306: expected MySQL, got %q", annotated[2].Label)
	}
}

func TestLabeler_Annotate_EmptySlice(t *testing.T) {
	l := NewLabeler(nil)
	result := l.Annotate([]ChangeEvent{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got len %d", len(result))
	}
}
