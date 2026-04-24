package portscanner

import (
	"testing"
	"time"
)

func makeEscalationEvent(port int, proto string, sev Severity) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:     port,
			Protocol: proto,
		},
		Severity: sev,
	}
}

func fixedEscalationNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestEscalator_BelowThreshold_NoEscalation(t *testing.T) {
	policy := EscalationPolicy{CountThreshold: 3, Window: time.Minute}
	e := NewEscalator(policy)
	base := time.Now()
	e.now = fixedEscalationNow(base)

	ev := makeEscalationEvent(8080, "tcp", SeverityLow)
	out := e.Process(ev)
	if out.Severity != SeverityLow {
		t.Errorf("expected Low, got %v", out.Severity)
	}
}

func TestEscalator_AtThreshold_Escalates(t *testing.T) {
	policy := EscalationPolicy{CountThreshold: 3, Window: time.Minute}
	e := NewEscalator(policy)
	base := time.Now()
	e.now = fixedEscalationNow(base)

	ev := makeEscalationEvent(8080, "tcp", SeverityLow)
	e.Process(ev)
	e.Process(ev)
	out := e.Process(ev)
	if out.Severity != SeverityHigh {
		t.Errorf("expected High after threshold, got %v", out.Severity)
	}
}

func TestEscalator_OldEventsEvicted_NoEscalation(t *testing.T) {
	policy := EscalationPolicy{CountThreshold: 3, Window: time.Minute}
	e := NewEscalator(policy)
	old := time.Now().Add(-2 * time.Minute)
	e.now = fixedEscalationNow(old)

	ev := makeEscalationEvent(9090, "tcp", SeverityMedium)
	e.Process(ev)
	e.Process(ev)

	// Advance time so old events are outside the window.
	e.now = fixedEscalationNow(time.Now())
	out := e.Process(ev)
	if out.Severity == SeverityHigh {
		t.Error("expected no escalation after window expired")
	}
}

func TestEscalator_IndependentPorts(t *testing.T) {
	policy := EscalationPolicy{CountThreshold: 2, Window: time.Minute}
	e := NewEscalator(policy)
	e.now = fixedEscalationNow(time.Now())

	ev80 := makeEscalationEvent(80, "tcp", SeverityLow)
	ev443 := makeEscalationEvent(443, "tcp", SeverityLow)

	e.Process(ev80)
	out80 := e.Process(ev80) // should escalate
	out443 := e.Process(ev443) // first time, should not escalate

	if out80.Severity != SeverityHigh {
		t.Errorf("expected port 80 to escalate, got %v", out80.Severity)
	}
	if out443.Severity == SeverityHigh {
		t.Error("expected port 443 not to escalate on first event")
	}
}

func TestEscalator_Flush_RemovesExpired(t *testing.T) {
	policy := EscalationPolicy{CountThreshold: 2, Window: time.Minute}
	e := NewEscalator(policy)
	old := time.Now().Add(-2 * time.Minute)
	e.now = fixedEscalationNow(old)

	ev := makeEscalationEvent(3000, "udp", SeverityLow)
	e.Process(ev)

	e.now = fixedEscalationNow(time.Now())
	e.Flush()

	if len(e.record) != 0 {
		t.Errorf("expected record to be empty after flush, got %d entries", len(e.record))
	}
}
