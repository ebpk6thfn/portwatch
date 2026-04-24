package portscanner

import (
	"testing"
	"time"
)

// TestCorrelator_PipelineStyleGrouping simulates a burst of port-open events
// flowing through the correlator as they would in a real pipeline tick.
func TestCorrelator_PipelineStyleGrouping(t *testing.T) {
	now := time.Now()
	policy := CorrelationPolicy{Window: 10 * time.Second, MinCount: 3}
	c := NewCorrelator(policy, func() time.Time { return now })

	events := []ChangeEvent{
		makeCorrelationEvent("tcp", "opened", 8080),
		makeCorrelationEvent("tcp", "opened", 8081),
		makeCorrelationEvent("tcp", "opened", 8082),
	}

	var group *CorrelatedGroup
	for _, e := range events {
		group = c.Add(e)
	}

	if group == nil {
		t.Fatal("expected correlated group after 3 events")
	}
	if len(group.Events) != 3 {
		t.Errorf("expected 3 correlated events, got %d", len(group.Events))
	}
	if group.Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %s", group.Protocol)
	}
	if group.FirstSeen.IsZero() || group.LastSeen.IsZero() {
		t.Error("expected non-zero timestamps")
	}
}

// TestCorrelator_MixedProtocols ensures tcp and udp events are correlated independently.
func TestCorrelator_MixedProtocols_Independent(t *testing.T) {
	now := time.Now()
	policy := CorrelationPolicy{Window: 10 * time.Second, MinCount: 2}
	c := NewCorrelator(policy, func() time.Time { return now })

	// One tcp, one udp — neither should trigger yet
	r1 := c.Add(makeCorrelationEvent("tcp", "opened", 8080))
	r2 := c.Add(makeCorrelationEvent("udp", "opened", 9090))
	if r1 != nil || r2 != nil {
		t.Error("expected no group before MinCount reached per protocol")
	}

	// Second tcp — should now trigger tcp group
	r3 := c.Add(makeCorrelationEvent("tcp", "opened", 8081))
	if r3 == nil {
		t.Error("expected tcp group after 2 tcp events")
	}
	if r3.Protocol != "tcp" {
		t.Errorf("expected tcp, got %s", r3.Protocol)
	}

	// udp group still pending — flush should return it
	groups := c.Flush()
	if len(groups) != 1 {
		t.Errorf("expected 1 pending udp group, got %d", len(groups))
	}
}
