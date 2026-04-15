package portscanner

import "testing"

func makeEnrichEvent(port uint16, process string, kind EventKind) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:    port,
			Proto:   "tcp",
			Address: "0.0.0.0",
			Process: process,
		},
		Kind: kind,
	}
}

func TestEnricher_SeverityAttached(t *testing.T) {
	c := NewClassifier(nil)
	e := NewEnricher(c, nil)

	events := []ChangeEvent{
		makeEnrichEvent(443, "nginx", EventOpened),
		makeEnrichEvent(54321, "", EventOpened),
	}

	enriched := e.Enrich(events)
	if len(enriched) != 2 {
		t.Fatalf("expected 2 enriched events, got %d", len(enriched))
	}
	if enriched[0].Severity != SeverityHigh {
		t.Errorf("port 443 should be high, got %s", enriched[0].Severity)
	}
	if enriched[1].Severity != SeverityLow {
		t.Errorf("ephemeral port should be low, got %s", enriched[1].Severity)
	}
}

func TestEnricher_LabelAttached(t *testing.T) {
	c := NewClassifier(nil)
	labels := map[uint16]string{80: "HTTP", 443: "HTTPS"}
	e := NewEnricher(c, labels)

	enriched := e.Enrich([]ChangeEvent{makeEnrichEvent(443, "caddy", EventOpened)})
	if len(enriched) != 1 {
		t.Fatal("expected 1 enriched event")
	}
	if enriched[0].Label != "HTTPS" {
		t.Errorf("expected label HTTPS, got %q", enriched[0].Label)
	}
}

func TestEnricher_NoLabel_EmptyString(t *testing.T) {
	c := NewClassifier(nil)
	e := NewEnricher(c, nil)
	enriched := e.Enrich([]ChangeEvent{makeEnrichEvent(9999, "app", EventOpened)})
	if enriched[0].Label != "" {
		t.Errorf("expected empty label, got %q", enriched[0].Label)
	}
}

func TestFilterBySeverity_KeepsHighOnly(t *testing.T) {
	c := NewClassifier(nil)
	e := NewEnricher(c, nil)
	events := []ChangeEvent{
		makeEnrichEvent(443, "nginx", EventOpened),
		makeEnrichEvent(8080, "app", EventOpened),
		makeEnrichEvent(54321, "", EventClosed),
	}
	enriched := e.Enrich(events)
	filtered := FilterBySeverity(enriched, SeverityHigh)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 high-severity event, got %d", len(filtered))
	}
	if filtered[0].Entry.Port != 443 {
		t.Errorf("expected port 443, got %d", filtered[0].Entry.Port)
	}
}

func TestFilterBySeverity_EmptyInput(t *testing.T) {
	result := FilterBySeverity(nil, SeverityLow)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}
