package portscanner

import (
	"testing"
	"time"
)

func makeJournalEntry(kind, protocol string, port uint16, severity string) JournalEntry {
	return JournalEntry{
		Timestamp: time.Now(),
		EventKey:  protocol + ":" + string(rune(port)),
		Kind:      kind,
		Protocol:  protocol,
		Port:      port,
		Process:   "testd",
		Severity:  severity,
		Note:      "test entry",
	}
}

func TestJournal_EmptyLen(t *testing.T) {
	j := NewJournal(DefaultJournalPolicy())
	if j.Len() != 0 {
		t.Fatalf("expected 0, got %d", j.Len())
	}
}

func TestJournal_RecordAndAll(t *testing.T) {
	j := NewJournal(DefaultJournalPolicy())
	j.Record(makeJournalEntry("opened", "tcp", 8080, "high"))
	j.Record(makeJournalEntry("closed", "udp", 53, "medium"))

	all := j.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Kind != "opened" {
		t.Errorf("expected opened, got %s", all[0].Kind)
	}
	if all[1].Protocol != "udp" {
		t.Errorf("expected udp, got %s", all[1].Protocol)
	}
}

func TestJournal_Overflow_EvictsOldest(t *testing.T) {
	policy := JournalPolicy{MaxEntries: 3}
	j := NewJournal(policy)

	for i := 0; i < 5; i++ {
		e := makeJournalEntry("opened", "tcp", uint16(8000+i), "low")
		e.Note = string(rune('A' + i))
		j.Record(e)
	}

	all := j.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", len(all))
	}
	// oldest (A, B) should be gone; C, D, E remain
	if all[0].Note != "C" {
		t.Errorf("expected oldest surviving entry note C, got %s", all[0].Note)
	}
}

func TestJournal_Clear(t *testing.T) {
	j := NewJournal(DefaultJournalPolicy())
	j.Record(makeJournalEntry("opened", "tcp", 443, "high"))
	j.Clear()
	if j.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", j.Len())
	}
}

func TestJournal_Since(t *testing.T) {
	j := NewJournal(DefaultJournalPolicy())

	old := time.Now().Add(-10 * time.Minute)
	recent := time.Now().Add(-1 * time.Minute)

	e1 := makeJournalEntry("opened", "tcp", 80, "high")
	e1.Timestamp = old
	e2 := makeJournalEntry("closed", "tcp", 443, "medium")
	e2.Timestamp = recent

	j.Record(e1)
	j.Record(e2)

	cutoff := time.Now().Add(-5 * time.Minute)
	result := j.Since(cutoff)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry since cutoff, got %d", len(result))
	}
	if result[0].Port != 443 {
		t.Errorf("expected port 443, got %d", result[0].Port)
	}
}

func TestJournal_DefaultPolicy_ZeroMaxUsesDefault(t *testing.T) {
	j := NewJournal(JournalPolicy{MaxEntries: 0})
	if j.policy.MaxEntries != DefaultJournalPolicy().MaxEntries {
		t.Errorf("expected default max entries, got %d", j.policy.MaxEntries)
	}
}
