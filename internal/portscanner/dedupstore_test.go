package portscanner

import (
	"testing"
	"time"
)

func TestDedupStore_FirstKey_NotSeen(t *testing.T) {
	ds := NewDedupStore(10 * time.Second)
	if ds.Seen("key1") {
		t.Fatal("expected first occurrence to not be seen")
	}
}

func TestDedupStore_SameKeyWithinWindow_IsSeen(t *testing.T) {
	ds := NewDedupStore(10 * time.Second)
	ds.Seen("key1")
	if !ds.Seen("key1") {
		t.Fatal("expected second occurrence within window to be seen")
	}
}

func TestDedupStore_SameKeyAfterWindow_NotSeen(t *testing.T) {
	now := time.Unix(1000, 0)
	ds := NewDedupStore(5 * time.Second)
	ds.now = func() time.Time { return now }

	ds.Seen("key1")

	ds.now = func() time.Time { return now.Add(6 * time.Second) }
	if ds.Seen("key1") {
		t.Fatal("expected key to not be seen after window expiry")
	}
}

func TestDedupStore_IndependentKeys(t *testing.T) {
	ds := NewDedupStore(10 * time.Second)
	ds.Seen("a")
	if ds.Seen("b") {
		t.Fatal("key b should not be seen after only key a was recorded")
	}
}

func TestDedupStore_Flush_RemovesExpired(t *testing.T) {
	now := time.Unix(1000, 0)
	ds := NewDedupStore(5 * time.Second)
	ds.now = func() time.Time { return now }

	ds.Seen("x")
	ds.Seen("y")

	ds.now = func() time.Time { return now.Add(6 * time.Second) }
	ds.Flush()

	if ds.Len() != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", ds.Len())
	}
}

func TestDedupStore_Len_CountsActive(t *testing.T) {
	now := time.Unix(1000, 0)
	ds := NewDedupStore(10 * time.Second)
	ds.now = func() time.Time { return now }

	ds.Seen("a")
	ds.Seen("b")
	ds.Seen("c")

	if ds.Len() != 3 {
		t.Fatalf("expected 3 active entries, got %d", ds.Len())
	}
}
