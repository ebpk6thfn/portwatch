package portscanner

import (
	"testing"
	"time"
)

func TestBurstAlert_NoAlertBelowThreshold(t *testing.T) {
	ba := NewBurstAlert(5, 10*time.Second, 30*time.Second, "tcp")
	for i := 0; i < 5; i++ {
		e := ChangeEvent{Entry: Entry{Protocol: "tcp", Port: uint16(8000 + i)}}
		if got := ba.Observe(e); got != nil {
			t.Errorf("unexpected alert on event %d", i+1)
		}
	}
}

func TestBurstAlert_AlertOnExceed(t *testing.T) {
	ba := NewBurstAlert(3, 10*time.Second, 30*time.Second, "tcp")
	var alert *ChangeEvent
	for i := 0; i < 4; i++ {
		e := ChangeEvent{Entry: Entry{Protocol: "tcp", Port: uint16(9000 + i)}}
		al)
	}
	if alert == nil {
		t.Fatal("expected burst alert")
	}
	if alert.Label != "burst-detected" {
		t.Errorf("expected label burst-detected, got %s", alert.Label)
	}
	if alert.Severity != SeverityHigh {
		t.Errorf("expected high severity")
	}
	if alert.Entry.Protocol != "tcp" {
		t.Errorf("expected tcp protocol in alert")
	}
}

func TestBurstAlert_CooldownSuppressesRepeat(t *testing.T) {
	ba := NewBurstAlert(1, 10*time.Second, 30*time.Second, "udp")
	e := ChangeEvent{Entry: Entry{Protocol: "udp", Port: 53}}
	// first burst
	ba.Observe(e)
	first := ba.Observe(e)
	if first == nil {
		t.Fatal("expected first burst alert")
	}
	// immediate repeat should be suppressed by cooldown
	second := ba.Observe(e)
	if second != nil {
		t.Error("expected cooldown to suppress second alert")
	}
}

func TestBurstAlert_ProtocolInAlert(t *testing.T) {
	ba := NewBurstAlert(1, 10*time.Second, 1*time.Millisecond, "udp")
	e := ChangeEvent{Entry: Entry{Protocol: "udp", Port: 514}}
	ba.Observe(e)
	alert := ba.Observe(e)
	if alert == nil {
		t.Fatal("expected alert")
	}
	if alert.Entry.Protocol != "udp" {
		t.Errorf("protocol mismatch: %s", alert.Entry.Protocol)
	}
}
