package portscanner

import (
	"testing"
	"time"
)

func makeAnomalyEvent(port uint16, proto string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{Port: port, Protocol: proto, Addr: "0.0.0.0"},
		Type:  EventOpened,
	}
}

func TestAnomalyDetector_NoBurst_NoAnomaly(t *testing.T) {
	d := NewAnomalyDetector(10*time.Second, 5, 30*time.Second, 60*time.Second)
	now := time.Now()
	ev := makeAnomalyEvent(8080, "tcp")
	result := d.Evaluate(ev, now)
	if result != nil {
		t.Fatalf("expected nil anomaly, got %v", result)
	}
}

func TestAnomalyDetector_BurstDetected(t *testing.T) {
	d := NewAnomalyDetector(10*time.Second, 3, 30*time.Second, 1*time.Millisecond)
	now := time.Now()
	ev := makeAnomalyEvent(9000, "tcp")

	var last *Anomaly
	for i := 0; i < 4; i++ {
		now = now.Add(100 * time.Millisecond)
		last = d.Evaluate(ev, now)
	}
	if last == nil {
		t.Fatal("expected burst anomaly")
	}
	if last.Type != AnomalyBurst {
		t.Fatalf("expected AnomalyBurst, got %s", last.Type)
	}
	if last.Port != 9000 {
		t.Fatalf("expected port 9000, got %d", last.Port)
	}
}

func TestAnomalyDetector_CooldownSuppressesRepeat(t *testing.T) {
	d := NewAnomalyDetector(10*time.Second, 2, 30*time.Second, 10*time.Minute)
	now := time.Now()
	ev := makeAnomalyEvent(443, "tcp")

	// trigger burst
	var anomalies []*Anomaly
	for i := 0; i < 5; i++ {
		now = now.Add(50 * time.Millisecond)
		a := d.Evaluate(ev, now)
		if a != nil {
			anomalies = append(anomalies, a)
		}
	}
	if len(anomalies) != 1 {
		t.Fatalf("expected exactly 1 anomaly due to cooldown, got %d", len(anomalies))
	}
}

func TestAnomaly_String(t *testing.T) {
	a := Anomaly{
		Type:       AnomalyBurst,
		Port:       80,
		Protocol:   "tcp",
		Score:      2.0,
		DetectedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Detail:     "4 events in window",
	}
	s := a.String()
	for _, want := range []string{"burst", "tcp/80", "2.00", "4 events"} {
		if !contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
