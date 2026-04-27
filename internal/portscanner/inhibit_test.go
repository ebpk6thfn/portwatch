package portscanner

import (
	"testing"
	"time"
)

func makeInhibitEvent(proto, ip string, port uint16) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Protocol:  proto,
			LocalIP:   ip,
			LocalPort: port,
		},
		Type: PortOpened,
	}
}

func TestInhibitor_NotInhibitedByDefault(t *testing.T) {
	policy := DefaultInhibitPolicy()
	inh := NewInhibitor(policy)
	if inh.IsInhibited("tcp:127.0.0.1:8080") {
		t.Fatal("expected key to not be inhibited initially")
	}
}

func TestInhibitor_InhibitBlocksKey(t *testing.T) {
	policy := InhibitPolicy{Duration: 1 * time.Minute}
	inh := NewInhibitor(policy)
	inh.Inhibit("tcp:127.0.0.1:8080")
	if !inh.IsInhibited("tcp:127.0.0.1:8080") {
		t.Fatal("expected key to be inhibited after Inhibit()")
	}
}

func TestInhibitor_ReleaseUnblocksKey(t *testing.T) {
	policy := InhibitPolicy{Duration: 1 * time.Minute}
	inh := NewInhibitor(policy)
	inh.Inhibit("tcp:127.0.0.1:9000")
	inh.Release("tcp:127.0.0.1:9000")
	if inh.IsInhibited("tcp:127.0.0.1:9000") {
		t.Fatal("expected key to be released")
	}
}

func TestInhibitor_ExpiresAfterDuration(t *testing.T) {
	now := time.Now()
	policy := InhibitPolicy{Duration: 30 * time.Second}
	inh := NewInhibitor(policy)
	inh.clock = func() time.Time { return now }
	inh.Inhibit("tcp:0.0.0.0:443")

	// advance past duration
	inh.clock = func() time.Time { return now.Add(31 * time.Second) }
	if inh.IsInhibited("tcp:0.0.0.0:443") {
		t.Fatal("expected inhibit to have expired")
	}
}

func TestInhibitor_Filter_DropsInhibitedEvents(t *testing.T) {
	policy := InhibitPolicy{Duration: 1 * time.Minute}
	inh := NewInhibitor(policy)

	ev1 := makeInhibitEvent("tcp", "0.0.0.0", 80)
	ev2 := makeInhibitEvent("tcp", "0.0.0.0", 443)

	inh.Inhibit(ev1.Entry.Key())

	result := inh.Filter([]ChangeEvent{ev1, ev2})
	if len(result) != 1 {
		t.Fatalf("expected 1 event after filter, got %d", len(result))
	}
	if result[0].Entry.LocalPort != 443 {
		t.Errorf("expected port 443 to pass, got %d", result[0].Entry.LocalPort)
	}
}

func TestInhibitor_Flush_RemovesExpired(t *testing.T) {
	now := time.Now()
	policy := InhibitPolicy{Duration: 10 * time.Second}
	inh := NewInhibitor(policy)
	inh.clock = func() time.Time { return now }

	inh.Inhibit("key-a")
	inh.Inhibit("key-b")

	inh.clock = func() time.Time { return now.Add(20 * time.Second) }
	inh.Flush()

	if inh.IsInhibited("key-a") || inh.IsInhibited("key-b") {
		t.Fatal("expected all inhibits to be flushed")
	}
}
