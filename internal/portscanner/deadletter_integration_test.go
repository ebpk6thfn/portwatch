package portscanner

import (
	"testing"
	"time"
)

// TestDeadLetterQueue_QuotaIntegration simulates a pipeline where quota
// exhaustion causes events to be routed to the dead-letter queue.
func TestDeadLetterQueue_QuotaIntegration(t *testing.T) {
	now := time.Now()
	policy := DefaultQuotaPolicy()
	policy.MaxPerWindow = 2
	policy.Window = 10 * time.Second

	quota := NewQuota(policy, func() time.Time { return now })
	dlq := NewDeadLetterQueue(16)

	events := []ChangeEvent{
		{Entry: Entry{LocalPort: 80, Protocol: "tcp"}, Type: EventOpened, Timestamp: now, Severity: SeverityHigh},
		{Entry: Entry{LocalPort: 443, Protocol: "tcp"}, Type: EventOpened, Timestamp: now, Severity: SeverityHigh},
		{Entry: Entry{LocalPort: 8080, Protocol: "tcp"}, Type: EventOpened, Timestamp: now, Severity: SeverityHigh},
	}

	var delivered []ChangeEvent
	for _, ev := range events {
		if quota.Allow(ev) {
			delivered = append(delivered, ev)
		} else {
			dlq.Push(ev, ReasonQuotaExceeded)
		}
	}

	if len(delivered) != 2 {
		t.Errorf("expected 2 delivered events, got %d", len(delivered))
	}
	if dlq.Len() != 1 {
		t.Errorf("expected 1 dead-letter entry, got %d", dlq.Len())
	}

	counts := dlq.CountByReason()
	if counts[ReasonQuotaExceeded] != 1 {
		t.Errorf("expected 1 quota_exceeded in dead-letter, got %d", counts[ReasonQuotaExceeded])
	}

	drained := dlq.Drain()
	if drained[0].Event.Entry.LocalPort != 8080 {
		t.Errorf("expected dead-lettered port 8080, got %d", drained[0].Event.Entry.LocalPort)
	}
}
