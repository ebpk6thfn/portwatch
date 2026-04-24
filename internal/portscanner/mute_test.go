package portscanner

import (
	"testing"
	"time"
)

func makeMuteNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestMuter_NotMutedByDefault(t *testing.T) {
	m := NewMuter(DefaultMutePolicy())
	if m.IsMuted("tcp:80") {
		t.Fatal("expected not muted by default")
	}
}

func TestMuter_MuteBlocksKey(t *testing.T) {
	m := NewMuter(MutePolicy{Duration: 5 * time.Minute})
	m.Mute("tcp:443")
	if !m.IsMuted("tcp:443") {
		t.Fatal("expected key to be muted")
	}
}

func TestMuter_UnmuteReleasesKey(t *testing.T) {
	m := NewMuter(MutePolicy{Duration: 5 * time.Minute})
	m.Mute("tcp:8080")
	m.Unmute("tcp:8080")
	if m.IsMuted("tcp:8080") {
		t.Fatal("expected key to be unmuted")
	}
}

func TestMuter_ExpiresAfterDuration(t *testing.T) {
	base := time.Now()
	m := NewMuter(MutePolicy{Duration: 1 * time.Minute})
	m.now = makeMuteNow(base)
	m.Mute("tcp:22")

	// Advance past expiry
	m.now = makeMuteNow(base.Add(2 * time.Minute))
	if m.IsMuted("tcp:22") {
		t.Fatal("expected mute to have expired")
	}
}

func TestMuter_IndependentKeys(t *testing.T) {
	m := NewMuter(MutePolicy{Duration: 5 * time.Minute})
	m.Mute("tcp:80")
	if !m.IsMuted("tcp:80") {
		t.Fatal("expected tcp:80 to be muted")
	}
	if m.IsMuted("tcp:443") {
		t.Fatal("expected tcp:443 to not be muted")
	}
}

func TestMuter_Filter_RemovesMuted(t *testing.T) {
	m := NewMuter(MutePolicy{Duration: 5 * time.Minute})

	e1 := makeEntry("tcp", "127.0.0.1", 80, "nginx")
	e2 := makeEntry("tcp", "127.0.0.1", 443, "nginx")
	events := []ChangeEvent{
		{Entry: e1, Type: EventOpened},
		{Entry: e2, Type: EventOpened},
	}

	m.Mute(e1.Key())
	out := m.Filter(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
	if out[0].Entry.Key() != e2.Key() {
		t.Fatalf("expected surviving event to be e2")
	}
}

func TestMuter_Filter_PassesAllWhenNoneMuted(t *testing.T) {
	m := NewMuter(MutePolicy{Duration: 5 * time.Minute})
	e1 := makeEntry("tcp", "0.0.0.0", 8080, "app")
	e2 := makeEntry("udp", "0.0.0.0", 53, "dns")
	events := []ChangeEvent{
		{Entry: e1, Type: EventOpened},
		{Entry: e2, Type: EventOpened},
	}
	out := m.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out))
	}
}
