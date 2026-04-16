package portscanner

import (
	"testing"
	"time"
)

func makeSnapForSampler(entries []Entry) *Snapshot {
	return NewSnapshot(entries, time.Now())
}

func TestSampler_EmptyLatest(t *testing.T) {
	s := NewSampler(10)
	_, ok := s.Latest()
	if ok {
		t.Fatal("expected no latest on empty sampler")
	}
}

func TestSampler_RecordAndLatest(t *testing.T) {
	s := NewSampler(10)
	snap := makeSnapForSampler([]Entry{
		makeSnapshotEntry("tcp", "0.0.0.0", 80),
		makeSnapshotEntry("tcp", "0.0.0.0", 443),
	})
	now := time.Now()
	s.Record(snap, now)

	rec, ok := s.Latest()
	if !ok {
		t.Fatal("expected latest after record")
	}
	if rec.Count != 2 {
		t.Fatalf("expected count 2, got %d", rec.Count)
	}
	if !rec.At.Equal(now) {
		t.Fatalf("unexpected timestamp")
	}
}

func TestSampler_LenTracking(t *testing.T) {
	s := NewSampler(10)
	snap := makeSnapForSampler(nil)
	for i := 0; i < 5; i++ {
		s.Record(snap, time.Now())
	}
	if s.Len() != 5 {
		t.Fatalf("expected len 5, got %d", s.Len())
	}
}

func TestSampler_Eviction(t *testing.T) {
	s := NewSampler(3)
	snap := makeSnapForSampler(nil)
	base := time.Now()
	for i := 0; i < 5; i++ {
		s.Record(snap, base.Add(time.Duration(i)*time.Second))
	}
	if s.Len() != 3 {
		t.Fatalf("expected len 3 after eviction, got %d", s.Len())
	}
	all := s.All()
	expected := base.Add(2 * time.Second)
	if !all[0].At.Equal(expected) {
		t.Fatalf("oldest retained sample mismatch")
	}
}

func TestSampler_AllReturnsCopy(t *testing.T) {
	s := NewSampler(10)
	snap := makeSnapForSampler(nil)
	s.Record(snap, time.Now())
	a := s.All()
	a[0].Count = 999
	b := s.All()
	if b[0].Count == 999 {
		t.Fatal("All() should return a copy")
	}
}
