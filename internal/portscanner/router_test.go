package portscanner

import (
	"net"
	"testing"
)

func makeRouterEvent(severity, protocol string, port uint16) ChangeEvent {
	return ChangeEvent{
		Severity: severity,
		Entry: Entry{
			Protocol:  protocol,
			LocalIP:   net.ParseIP("127.0.0.1"),
			LocalPort: port,
		},
	}
}

func TestRouter_DefaultLabel_WhenNoRules(t *testing.T) {
	r := NewRouter("default", nil)
	e := makeRouterEvent("low", "tcp", 8080)
	if got := r.Route(e); got != "default" {
		t.Fatalf("expected default, got %s", got)
	}
}

func TestRouter_MatchesSeverity(t *testing.T) {
	rules := []RouteRule{{Severity: "high", DestLabel: "pagerduty"}}
	r := NewRouter("default", rules)
	e := makeRouterEvent("high", "tcp", 22)
	if got := r.Route(e); got != "pagerduty" {
		t.Fatalf("expected pagerduty, got %s", got)
	}
}

func TestRouter_NoMatchFallsToDefault(t *testing.T) {
	rules := []RouteRule{{Severity: "high", DestLabel: "pagerduty"}}
	r := NewRouter("default", rules)
	e := makeRouterEvent("low", "tcp", 9090)
	if got := r.Route(e); got != "default" {
		t.Fatalf("expected default, got %s", got)
	}
}

func TestRouter_MatchesProtocol(t *testing.T) {
	rules := []RouteRule{{Protocol: "udp", DestLabel: "udp-sink"}}
	r := NewRouter("default", rules)
	e := makeRouterEvent("low", "udp", 53)
	if got := r.Route(e); got != "udp-sink" {
		t.Fatalf("expected udp-sink, got %s", got)
	}
}

func TestRouter_RouteAll_GroupsByLabel(t *testing.T) {
	rules := []RouteRule{{Severity: "high", DestLabel: "alerts"}}
	r := NewRouter("default", rules)
	events := []ChangeEvent{
		makeRouterEvent("high", "tcp", 22),
		makeRouterEvent("low", "tcp", 8080),
		makeRouterEvent("high", "udp", 53),
	}
	result := r.RouteAll(events)
	if len(result["alerts"]) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(result["alerts"]))
	}
	if len(result["default"]) != 1 {
		t.Fatalf("expected 1 default, got %d", len(result["default"]))
	}
}

func TestRouter_RouteAll_EmptyEvents(t *testing.T) {
	r := NewRouter("default", nil)
	result := r.RouteAll(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}
