package portscanner

import (
	"strings"
	"testing"
)

func TestEntry_Key_Uniqueness(t *testing.T) {
	a := Entry{LocalAddr: "0.0.0.0", Port: 80, Protocol: ProtoTCP}
	b := Entry{LocalAddr: "127.0.0.1", Port: 80, Protocol: ProtoTCP}
	c := Entry{LocalAddr: "0.0.0.0", Port: 80, Protocol: ProtoUDP}
	d := Entry{LocalAddr: "0.0.0.0", Port: 80, Protocol: ProtoTCP}

	if a.Key() == b.Key() {
		t.Error("entries with different local addrs should have different keys")
	}
	if a.Key() == c.Key() {
		t.Error("entries with different protocols should have different keys")
	}
	if a.Key() != d.Key() {
		t.Error("identical entries should have the same key")
	}
}

func TestEntry_Key_Format(t *testing.T) {
	e := Entry{LocalAddr: "0.0.0.0", Port: 8080, Protocol: ProtoTCP}
	key := e.Key()
	if !strings.Contains(key, "tcp") {
		t.Errorf("key %q should contain protocol", key)
	}
	if !strings.Contains(key, "8080") {
		t.Errorf("key %q should contain port", key)
	}
	if !strings.Contains(key, "0.0.0.0") {
		t.Errorf("key %q should contain local addr", key)
	}
}

func TestEntry_String_WithProcess(t *testing.T) {
	e := Entry{LocalAddr: "127.0.0.1", Port: 3000, Protocol: ProtoTCP, PID: 42, ProcessName: "myapp"}
	s := e.String()
	if !strings.Contains(s, "myapp") {
		t.Errorf("String() %q should contain process name", s)
	}
	if !strings.Contains(s, "42") {
		t.Errorf("String() %q should contain PID", s)
	}
}

func TestEntry_String_WithoutProcess(t *testing.T) {
	e := Entry{LocalAddr: "0.0.0.0", Port: 80, Protocol: ProtoTCP}
	s := e.String()
	if !strings.Contains(s, "80") {
		t.Errorf("String() %q should contain port", s)
	}
	if strings.Contains(s, "pid") {
		t.Errorf("String() %q should not contain pid when ProcessName is empty", s)
	}
}

func TestChangeEvent_Fields(t *testing.T) {
	entry := Entry{LocalAddr: "0.0.0.0", Port: 9090, Protocol: ProtoTCP}
	event := ChangeEvent{Type: "opened", Entry: entry}

	if event.Type != "opened" {
		t.Errorf("expected type 'opened', got %q", event.Type)
	}
	if event.Entry.Port != 9090 {
		t.Errorf("expected port 9090, got %d", event.Entry.Port)
	}
}
