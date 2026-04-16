package portscanner

import (
	"testing"
)

func makeTaggerEvent(port uint16, proto string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:     port,
			Protocol: proto,
		},
	}
}

func containsTag(tags []Tag, target Tag) bool {
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}

func TestTagger_WellKnownPort(t *testing.T) {
	tgr := NewTagger(nil)
	tags := tgr.Tag(makeTaggerEvent(443, "tcp"))
	if !containsTag(tags, TagWellKnown) {
		t.Errorf("expected TagWellKnown for port 443, got %v", tags)
	}
	if !containsTag(tags, TagPrivileged) {
		t.Errorf("expected TagPrivileged for port 443, got %v", tags)
	}
}

func TestTagger_PrivilegedUnknown(t *testing.T) {
	tgr := NewTagger(nil)
	tags := tgr.Tag(makeTaggerEvent(999, "tcp"))
	if !containsTag(tags, TagPrivileged) {
		t.Errorf("expected TagPrivileged for port 999")
	}
	if containsTag(tags, TagWellKnown) {
		t.Errorf("did not expect TagWellKnown for port 999")
	}
}

func TestTagger_EphemeralPort(t *testing.T) {
	tgr := NewTagger(nil)
	tags := tgr.Tag(makeTaggerEvent(51000, "tcp"))
	if !containsTag(tags, TagEphemeral) {
		t.Errorf("expected TagEphemeral for port 51000, got %v", tags)
	}
}

func TestTagger_UserDefinedPort(t *testing.T) {
	tgr := NewTagger([]uint16{8080})
	tags := tgr.Tag(makeTaggerEvent(8080, "tcp"))
	if !containsTag(tags, TagUserDefined) {
		t.Errorf("expected TagUserDefined for port 8080, got %v", tags)
	}
}

func TestTagger_TagAll(t *testing.T) {
	tgr := NewTagger(nil)
	events := []ChangeEvent{
		makeTaggerEvent(80, "tcp"),
		makeTaggerEvent(8080, "tcp"),
		makeTaggerEvent(51000, "tcp"),
	}
	out := tgr.TagAll(events)
	if !containsTag(out[0], TagWellKnown) {
		t.Errorf("index 0: expected TagWellKnown")
	}
	if containsTag(out[1], TagEphemeral) {
		t.Errorf("index 1: port 8080 should not be ephemeral")
	}
	if !containsTag(out[2], TagEphemeral) {
		t.Errorf("index 2: expected TagEphemeral")
	}
}
