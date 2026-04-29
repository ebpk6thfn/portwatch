package portscanner

import (
	"testing"
	"time"
)

func markerNow(base time.Time) func() time.Time {
	t := base
	return func() time.Time { return t }
}

func TestMarker_GetUnmarked_ReturnsFalse(t *testing.T) {
	m := NewMarker(DefaultMarkerPolicy())
	_, ok := m.Get("key1")
	if ok {
		t.Fatal("expected no mark for unknown key")
	}
}

func TestMarker_MarkAndGet_ReturnsLabel(t *testing.T) {
	m := NewMarker(DefaultMarkerPolicy())
	m.Mark("key1", "needs-review")
	label, ok := m.Get("key1")
	if !ok {
		t.Fatal("expected mark to exist")
	}
	if label != "needs-review" {
		t.Fatalf("expected 'needs-review', got %q", label)
	}
}

func TestMarker_Unmark_RemovesMark(t *testing.T) {
	m := NewMarker(DefaultMarkerPolicy())
	m.Mark("key1", "label")
	m.Unmark("key1")
	_, ok := m.Get("key1")
	if ok {
		t.Fatal("expected mark to be removed after Unmark")
	}
}

func TestMarker_ExpiredMark_ReturnsFalse(t *testing.T) {
	base := time.Now()
	clockFn := markerNow(base)
	policy := MarkerPolicy{TTL: 5 * time.Minute}
	m := NewMarker(policy)
	m.now = clockFn
	m.Mark("key1", "stale")

	// advance clock past TTL
	m.now = func() time.Time { return base.Add(6 * time.Minute) }

	_, ok := m.Get("key1")
	if ok {
		t.Fatal("expected expired mark to return false")
	}
}

func TestMarker_ActiveMarkWithinTTL_ReturnsTrue(t *testing.T) {
	base := time.Now()
	policy := MarkerPolicy{TTL: 10 * time.Minute}
	m := NewMarker(policy)
	m.now = func() time.Time { return base }
	m.Mark("key1", "active")

	m.now = func() time.Time { return base.Add(9 * time.Minute) }

	_, ok := m.Get("key1")
	if !ok {
		t.Fatal("expected mark to still be active within TTL")
	}
}

func TestMarker_Flush_RemovesExpired(t *testing.T) {
	base := time.Now()
	policy := MarkerPolicy{TTL: 5 * time.Minute}
	m := NewMarker(policy)
	m.now = func() time.Time { return base }
	m.Mark("k1", "a")
	m.Mark("k2", "b")

	m.now = func() time.Time { return base.Add(6 * time.Minute) }
	m.Flush()

	if m.Len() != 0 {
		t.Fatalf("expected 0 marks after flush, got %d", m.Len())
	}
}

func TestMarker_ZeroTTL_NeverExpires(t *testing.T) {
	base := time.Now()
	policy := MarkerPolicy{TTL: 0}
	m := NewMarker(policy)
	m.now = func() time.Time { return base }
	m.Mark("key1", "permanent")

	m.now = func() time.Time { return base.Add(999 * time.Hour) }

	_, ok := m.Get("key1")
	if !ok {
		t.Fatal("expected zero-TTL mark to never expire")
	}
}
