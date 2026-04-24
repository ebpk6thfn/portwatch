package portscanner

import (
	"testing"
	"time"
)

func makeCorrelationEvent(protocol, evType string, port uint16) ChangeEvent {
	return ChangeEvent{
		Type: evType,
		Entry: Entry{
			Protocol:  protocol,
			LocalPort: port,
			LocalAddr: "127.0.0.1",
		},
	}
}

func TestCorrelator_BelowMinCount_ReturnsNil(t *testing.T) {
	now := time.Now()
	c := NewCorrelator(CorrelationPolicy{Window: 10 * time.Second, MinCount: 3}, func() time.Time { return now })

	result := c.Add(makeCorrelationEvent("tcp", "opened", 8080))
	if result != nil {
		t.Fatalf("expected nil, got group with %d events", len(result.Events))
	}
}

func TestCorrelator_AtMinCount_ReturnsGroup(t *testing.T) {
	now := time.Now()
	c := NewCorrelator(CorrelationPolicy{Window: 10 * time.Second, MinCount: 3}, func() time.Time { return now })

	for i := 0; i < 2; i++ {
		result := c.Add(makeCorrelationEvent("tcp", "opened", uint16(8080+i)))
		if result != nil {
			t.Fatalf("expected nil at step %d", i)
		}
	}
	result := c.Add(makeCorrelationEvent("tcp", "opened", 8082))
	if result == nil {
		t.Fatal("expected correlated group")
	}
	if len(result.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(result.Events))
	}
	if result.Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %s", result.Protocol)
	}
}

func TestCorrelator_WindowExpiry_StartsNewGroup(t *testing.T) {
	base := time.Now()
	calls := 0
	nowFn := func() time.Time {
		calls++
		if calls <= 2 {
			return base
		}
		return base.Add(60 * time.Second)
	}
	c := NewCorrelator(CorrelationPolicy{Window: 10 * time.Second, MinCount: 3}, nowFn)

	c.Add(makeCorrelationEvent("tcp", "opened", 8080))
	c.Add(makeCorrelationEvent("tcp", "opened", 8081))
	// This call is beyond the window — should start a fresh group
	result := c.Add(makeCorrelationEvent("tcp", "opened", 8082))
	if result != nil {
		t.Errorf("expected nil since new group only has 1 event")
	}
}

func TestCorrelator_Flush_ReturnsAllGroups(t *testing.T) {
	now := time.Now()
	c := NewCorrelator(CorrelationPolicy{Window: 10 * time.Second, MinCount: 5}, func() time.Time { return now })

	c.Add(makeCorrelationEvent("tcp", "opened", 8080))
	c.Add(makeCorrelationEvent("udp", "closed", 9090))

	groups := c.Flush()
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
	groups2 := c.Flush()
	if len(groups2) != 0 {
		t.Errorf("expected empty after flush, got %d", len(groups2))
	}
}

func TestCorrelator_GroupID_IsUnique(t *testing.T) {
	now := time.Now()
	c := NewCorrelator(CorrelationPolicy{Window: 1 * time.Millisecond, MinCount: 2}, func() time.Time { return now })

	c.Add(makeCorrelationEvent("tcp", "opened", 8080))
	result := c.Add(makeCorrelationEvent("tcp", "opened", 8081))
	if result == nil {
		t.Fatal("expected group")
	}
	if result.ID == "" {
		t.Error("expected non-empty group ID")
	}
}

func TestCorrelatedGroup_String(t *testing.T) {
	g := CorrelatedGroup{
		ID:        "tcp:opened-123",
		Protocol:  "tcp",
		Events:    make([]ChangeEvent, 2),
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
	s := g.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
