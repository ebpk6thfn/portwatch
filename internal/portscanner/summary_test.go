package portscanner

import (
	"strings"
	"testing"
	"time"
)

func makeSummaryEvent(proto string, port uint16, typ ChangeType) ChangeEvent {
	return ChangeEvent{
		Type: typ,
		Entry: Entry{Protocol: proto, Port: port},
	}
}

func fixedSummaryNow() func() time.Time {
	t := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func TestSummaryBuilder_EmptyBuild(t *testing.T) {
	sb := NewSummaryBuilder(time.Minute, fixedSummaryNow())
	r := sb.Build()
	if r.Opened != 0 || r.Closed != 0 || r.Suppressed != 0 || r.Anomalies != 0 {
		t.Fatalf("expected all zeros, got %+v", r)
	}
	if len(r.TopPorts) != 0 {
		t.Fatalf("expected no top ports")
	}
}

func TestSummaryBuilder_CountsOpenedClosed(t *testing.T) {
	sb := NewSummaryBuilder(time.Minute, fixedSummaryNow())
	sb.Record(makeSummaryEvent("tcp", 80, EventOpened))
	sb.Record(makeSummaryEvent("tcp", 443, EventOpened))
	sb.Record(makeSummaryEvent("tcp", 22, EventClosed))
	r := sb.Build()
	if r.Opened != 2 {
		t.Fatalf("expected 2 opened, got %d", r.Opened)
	}
	if r.Closed != 1 {
		t.Fatalf("expected 1 closed, got %d", r.Closed)
	}
}

func TestSummaryBuilder_SuppressedAndAnomalies(t *testing.T) {
	sb := NewSummaryBuilder(time.Minute, fixedSummaryNow())
	sb.RecordSuppressed()
	sb.RecordSuppressed()
	sb.RecordAnomaly()
	r := sb.Build()
	if r.Suppressed != 2 {
		t.Fatalf("expected 2 suppressed, got %d", r.Suppressed)
	}
	if r.Anomalies != 1 {
		t.Fatalf("expected 1 anomaly, got %d", r.Anomalies)
	}
}

func TestSummaryBuilder_ResetAfterBuild(t *testing.T) {
	sb := NewSummaryBuilder(time.Minute, fixedSummaryNow())
	sb.Record(makeSummaryEvent("tcp", 80, EventOpened))
	sb.RecordSuppressed()
	sb.Build()
	r2 := sb.Build()
	if r2.Opened != 0 || r2.Suppressed != 0 {
		t.Fatalf("expected reset after build, got %+v", r2)
	}
}

func TestSummaryBuilder_TopPorts(t *testing.T) {
	sb := NewSummaryBuilder(time.Minute, fixedSummaryNow())
	for i := 0; i < 3; i++ {
		sb.Record(makeSummaryEvent("tcp", 80, EventOpened))
	}
	sb.Record(makeSummaryEvent("tcp", 443, EventOpened))
	r := sb.Build()
	if len(r.TopPorts) == 0 {
		t.Fatal("expected top ports")
	}
	if r.TopPorts[0] != "tcp/80" {
		t.Fatalf("expected tcp/80 at top, got %s", r.TopPorts[0])
	}
}

func TestSummaryReport_String(t *testing.T) {
	sb := NewSummaryBuilder(time.Minute, fixedSummaryNow())
	sb.Record(makeSummaryEvent("tcp", 80, EventOpened))
	sb.RecordAnomaly()
	r := sb.Build()
	s := r.String()
	if !strings.Contains(s, "opened=1") {
		t.Fatalf("expected opened=1 in string: %s", s)
	}
	if !strings.Contains(s, "anomalies=1") {
		t.Fatalf("expected anomalies=1 in string: %s", s)
	}
}
