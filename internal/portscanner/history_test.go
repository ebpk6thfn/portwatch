package portscanner

import (
	"testing"
	"time"
)

func snap(n int) Snapshot {
	return Snapshot{
		CapturedAt:   time.Unix(int64(n), 0),
		ScanDuration: time.Duration(n) * time.Millisecond,
	}
}

func TestHistory_EmptyLatest(t *testing.T) {
	h := NewHistory(3)
	_, ok := h.Latest()
	if ok {
		t.Error("expected false for empty history")
	}
}

func TestHistory_AddAndLatest(t *testing.T) {
	h := NewHistory(3)
	h.Add(snap(1))
	h.Add(snap(2))

	s, ok := h.Latest()
	if !ok {
		t.Fatal("expected snapshot, got none")
	}
	if s.ScanDuration != 2*time.Millisecond {
		t.Errorf("expected latest to be snap(2), got %v", s.ScanDuration)
	}
}

func TestHistory_LenTracking(t *testing.T) {
	h := NewHistory(5)
	for i := 0; i < 3; i++ {
		h.Add(snap(i))
	}
	if h.Len() != 3 {
		t.Errorf("expected Len 3, got %d", h.Len())
	}
}

func TestHistory_Eviction(t *testing.T) {
	h := NewHistory(3)
	for i := 1; i <= 5; i++ {
		h.Add(snap(i))
	}
	if h.Len() != 3 {
		t.Errorf("expected Len 3 after eviction, got %d", h.Len())
	}
	all := h.All()
	// Should contain snaps 3, 4, 5
	expected := []int{3, 4, 5}
	for i, s := range all {
		if s.ScanDuration != time.Duration(expected[i])*time.Millisecond {
			t.Errorf("pos %d: expected snap(%d), got %v", i, expected[i], s.ScanDuration)
		}
	}
}

func TestHistory_AllChronologicalOrder(t *testing.T) {
	h := NewHistory(4)
	for i := 1; i <= 4; i++ {
		h.Add(snap(i))
	}
	all := h.All()
	for i := 1; i < len(all); i++ {
		if all[i].CapturedAt.Before(all[i-1].CapturedAt) {
			t.Errorf("shots not in chronological order at index %d", i)
		}
	}
}

func TestNewHistory_MinCapOne(t *testing.T) {
	h := NewHistory(0)
	h.Add(snap(1))
	h.Add(snap(2))
	if h.Len() != 1 {
		t.Errorf("expected Len 1 for cap-1 history, got %d", h.Len())
	}
	s, _ := h.Latest()
	if s.ScanDuration != 2*time.Millisecond {
		t.Errorf("expected latest snap(2), got %v", s.ScanDuration)
	}
}
