package portscanner

import (
	"testing"
	"time"
)

func makeRetentionEvent(port uint16) ChangeEvent {
	return ChangeEvent{
		Type:  EventOpened,
		Entry: Entry{Port: port, Protocol: "tcp"},
	}
}

func TestRetention_AddAndAll(t *testing.T) {
	rs := NewRetentionStore(RetentionPolicy{MaxCount: 10})
	rs.Add(makeRetentionEvent(80))
	rs.Add(makeRetentionEvent(443))
	if rs.Len() != 2 {
		t.Fatalf("expected 2, got %d", rs.Len())
	}
}

func TestRetention_MaxCount_Evicts(t *testing.T) {
	rs := NewRetentionStore(RetentionPolicy{MaxCount: 3})
	for i := uint16(1); i <= 5; i++ {
		rs.Add(makeRetentionEvent(i))
	}
	if rs.Len() != 3 {
		t.Fatalf("expected 3, got %d", rs.Len())
	}
	events := rs.All()
	if events[0].Entry.Port != 3 {
		t.Errorf("expected oldest kept port=3, got %d", events[0].Entry.Port)
	}
}

func TestRetention_MaxAge_Evicts(t *testing.T) {
	now := time.Now()
	rs := NewRetentionStore(RetentionPolicy{MaxAge: 5 * time.Second})
	rs.now = func() time.Time { return now }
	rs.Add(makeRetentionEvent(80))

	rs.now = func() time.Time { return now.Add(10 * time.Second) }
	rs.Add(makeRetentionEvent(443))

	if rs.Len() != 1 {
		t.Fatalf("expected 1 after age eviction, got %d", rs.Len())
	}
	if rs.All()[0].Entry.Port != 443 {
		t.Errorf("expected port 443 to remain")
	}
}

func TestRetention_Empty(t *testing.T) {
	rs := NewRetentionStore(RetentionPolicy{})
	if rs.Len() != 0 {
		t.Errorf("expected empty store")
	}
	if len(rs.All()) != 0 {
		t.Errorf("expected empty slice")
	}
}

func TestRetention_MaxCountAndAge_Combined(t *testing.T) {
	now := time.Now()
	rs := NewRetentionStore(RetentionPolicy{MaxAge: 10 * time.Second, MaxCount: 2})
	rs.now = func() time.Time { return now }
	rs.Add(makeRetentionEvent(1))
	rs.Add(makeRetentionEvent(2))
	rs.Add(makeRetentionEvent(3))

	// Only 2 kept by count
	if rs.Len() != 2 {
		t.Fatalf("expected 2, got %d", rs.Len())
	}

	// Advance time to evict by age
	rs.now = func() time.Time { return now.Add(20 * time.Second) }
	rs.Add(makeRetentionEvent(99))
	if rs.Len() != 1 {
		t.Fatalf("expected 1 after age eviction, got %d", rs.Len())
	}
}
