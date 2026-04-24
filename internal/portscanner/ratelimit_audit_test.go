package portscanner

import (
	"testing"
	"time"
)

var auditNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func TestRateLimitAuditLog_InitialEmpty(t *testing.T) {
	log := NewRateLimitAuditLog(10)
	if log.Len() != 0 {
		t.Fatalf("expected 0, got %d", log.Len())
	}
	if len(log.All()) != 0 {
		t.Fatal("expected empty slice")
	}
}

func TestRateLimitAuditLog_RecordAndAll(t *testing.T) {
	log := NewRateLimitAuditLog(10)
	log.Record("tcp:80", true, "first-event", auditNow)
	log.Record("tcp:80", false, "cooldown", auditNow.Add(time.Second))

	entries := log.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "tcp:80" || !entries[0].Allowed {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Reason != "cooldown" || entries[1].Allowed {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestRateLimitAuditLog_Overflow_EvictsOldest(t *testing.T) {
	log := NewRateLimitAuditLog(3)
	for i := 0; i < 4; i++ {
		log.Record("key", true, "r", auditNow.Add(time.Duration(i)*time.Second))
	}
	if log.Len() != 3 {
		t.Fatalf("expected 3, got %d", log.Len())
	}
	// Oldest entry (i=0) should have been evicted; earliest remaining is i=1.
	entries := log.All()
	if entries[0].Timestamp != auditNow.Add(time.Second) {
		t.Errorf("expected oldest evicted, got %v", entries[0].Timestamp)
	}
}

func TestRateLimitAuditLog_Clear(t *testing.T) {
	log := NewRateLimitAuditLog(10)
	log.Record("k", true, "r", auditNow)
	log.Clear()
	if log.Len() != 0 {
		t.Fatal("expected empty after Clear")
	}
}

func TestRateLimitAuditLog_DefaultMaxSize(t *testing.T) {
	log := NewRateLimitAuditLog(0)
	// Fill beyond default (256) to confirm cap is applied.
	for i := 0; i < 300; i++ {
		log.Record("k", true, "r", auditNow)
	}
	if log.Len() != 256 {
		t.Fatalf("expected 256, got %d", log.Len())
	}
}

func TestRateLimitAuditEntry_String_Allowed(t *testing.T) {
	e := RateLimitAuditEntry{
		Key:       "udp:53",
		Allowed:   true,
		Reason:    "first-event",
		Timestamp: auditNow,
	}
	s := e.String()
	for _, want := range []string{"ALLOWED", "udp:53", "first-event"} {
		if !containsSubstr(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func TestRateLimitAuditEntry_String_Suppressed(t *testing.T) {
	e := RateLimitAuditEntry{
		Key:       "tcp:443",
		Allowed:   false,
		Reason:    "cooldown",
		Timestamp: auditNow,
	}
	s := e.String()
	if !containsSubstr(s, "SUPPRESSED") {
		t.Errorf("String() missing SUPPRESSED in %q", s)
	}
}

// containsSubstr is a local helper to avoid importing strings in tests.
func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
