package portscanner

import (
	"net"
	"testing"
)

func makeWLEntry(port uint16, proto string) Entry {
	return Entry{
		Port:     port,
		Protocol: proto,
		LocalIP:  net.ParseIP("0.0.0.0"),
	}
}

func makeWLEvent(port uint16, kind ChangeKind) ChangeEvent {
	return ChangeEvent{
		Entry: makeWLEntry(port, "tcp"),
		Kind:  kind,
	}
}

func TestWatchlist_NoRules_PassesAll(t *testing.T) {
	wl := NewWatchlist(nil)
	events := []ChangeEvent{makeWLEvent(80, PortOpened), makeWLEvent(443, PortClosed)}
	out := wl.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out))
	}
}

func TestWatchlist_IgnoreRule_DropsEvent(t *testing.T) {
	wl := NewWatchlist([]WatchlistRule{
		{Port: 8080, Action: ActionIgnore},
	})
	events := []ChangeEvent{makeWLEvent(8080, PortOpened), makeWLEvent(443, PortOpened)}
	out := wl.Filter(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
	if out[0].Entry.Port != 443 {
		t.Errorf("expected port 443, got %d", out[0].Entry.Port)
	}
}

func TestWatchlist_AlertRule_KeepsEvent(t *testing.T) {
	wl := NewWatchlist([]WatchlistRule{
		{Port: 22, Action: ActionAlert},
	})
	events := []ChangeEvent{makeWLEvent(22, PortOpened)}
	out := wl.Filter(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
}

func TestWatchlist_Add_OverridesExisting(t *testing.T) {
	wl := NewWatchlist([]WatchlistRule{
		{Port: 9000, Action: ActionAlert},
	})
	wl.Add(9000, ActionIgnore)
	action, ok := wl.Evaluate(makeWLEvent(9000, PortOpened))
	if !ok {
		t.Fatal("expected rule to exist")
	}
	if action != ActionIgnore {
		t.Errorf("expected ActionIgnore, got %s", action)
	}
}

func TestWatchlist_Remove_DeletesRule(t *testing.T) {
	wl := NewWatchlist([]WatchlistRule{
		{Port: 3306, Action: ActionIgnore},
	})
	wl.Remove(3306)
	_, ok := wl.Evaluate(makeWLEvent(3306, PortOpened))
	if ok {
		t.Error("expected rule to be removed")
	}
}

func TestWatchlist_Evaluate_NoMatch(t *testing.T) {
	wl := NewWatchlist(nil)
	_, ok := wl.Evaluate(makeWLEvent(1234, PortOpened))
	if ok {
		t.Error("expected no match for unknown port")
	}
}
