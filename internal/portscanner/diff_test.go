package portscanner

import (
	"testing"
)

func makeEntry(port int, proto string) PortEntry {
	return PortEntry{
		Protocol:  proto,
		LocalAddr: "00000000:" + string(rune(port)),
		Port:      port,
		State:     "0A",
	}
}

func TestDiff_NoChanges(t *testing.T) {
	snap := map[int]PortEntry{
		80:  makeEntry(80, "tcp"),
		443: makeEntry(443, "tcp"),
	}
	changes := Diff(snap, snap)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestDiff_PortOpened(t *testing.T) {
	prev := map[int]PortEntry{
		80: makeEntry(80, "tcp"),
	}
	curr := map[int]PortEntry{
		80:   makeEntry(80, "tcp"),
		8080: makeEntry(8080, "tcp"),
	}
	changes := Diff(prev, curr)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Type != PortOpened {
		t.Errorf("expected PortOpened, got %s", changes[0].Type)
	}
	if changes[0].Entry.Port != 8080 {
		t.Errorf("expected port 8080, got %d", changes[0].Entry.Port)
	}
}

func TestDiff_PortClosed(t *testing.T) {
	prev := map[int]PortEntry{
		80:   makeEntry(80, "tcp"),
		9000: makeEntry(9000, "tcp"),
	}
	curr := map[int]PortEntry{
		80: makeEntry(80, "tcp"),
	}
	changes := Diff(prev, curr)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Type != PortClosed {
		t.Errorf("expected PortClosed, got %s", changes[0].Type)
	}
	if changes[0].Entry.Port != 9000 {
		t.Errorf("expected port 9000, got %d", changes[0].Entry.Port)
	}
}

func TestDiff_MultipleChanges(t *testing.T) {
	prev := map[int]PortEntry{
		22:  makeEntry(22, "tcp"),
		80:  makeEntry(80, "tcp"),
	}
	curr := map[int]PortEntry{
		22:   makeEntry(22, "tcp"),
		443:  makeEntry(443, "tcp"),
		8443: makeEntry(8443, "tcp"),
	}
	changes := Diff(prev, curr)

	opened, closed := 0, 0
	for _, c := range changes {
		switch c.Type {
		case PortOpened:
			opened++
		case PortClosed:
			closed++
		}
	}
	if opened != 2 {
		t.Errorf("expected 2 opened, got %d", opened)
	}
	if closed != 1 {
		t.Errorf("expected 1 closed, got %d", closed)
	}
}
