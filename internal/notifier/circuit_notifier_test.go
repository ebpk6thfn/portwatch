package notifier

import (
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

type stubNotifier struct {
	calls int
	err   error
}

func (s *stubNotifier) Notify(_ portscanner.ChangeEvent) error {
	s.calls++
	return s.err
}

func makeCircuitEvent() portscanner.ChangeEvent {
	return portscanner.ChangeEvent{Port: 8080, Protocol: "tcp", Type: portscanner.Opened}
}

func TestCircuitNotifier_AllowsWhenClosed(t *testing.T) {
	stub := &stubNotifier{}
	cn := NewCircuitNotifier(stub, 3, time.Second)
	if err := cn.Notify(makeCircuitEvent()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestCircuitNotifier_TripsOnFailures(t *testing.T) {
	stub := &stubNotifier{err: errors.New("boom")}
	cn := NewCircuitNotifier(stub, 2, time.Second)
	cn.Notify(makeCircuitEvent())
	cn.Notify(makeCircuitEvent())
	if cn.State() != portscanner.CircuitOpen {
		t.Fatal("expected circuit open")
	}
	err := cn.Notify(makeCircuitEvent())
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if stub.calls != 2 {
		t.Fatalf("inner should only have been called 2 times, got %d", stub.calls)
	}
}

func TestCircuitNotifier_RecoveryOnSuccess(t *testing.T) {
	stub := &stubNotifier{err: errors.New("fail")}
	cn := NewCircuitNotifier(stub, 1, time.Second)
	cn.Notify(makeCircuitEvent())
	if cn.State() != portscanner.CircuitOpen {
		t.Fatal("expected open")
	}
	stub.err = nil
	cn.circuit.RecordSuccess() // simulate recovery probe
	if err := cn.Notify(makeCircuitEvent()); err != nil {
		t.Fatalf("unexpected error after recovery: %v", err)
	}
	if cn.State() != portscanner.CircuitClosed {
		t.Fatal("expected closed after success")
	}
}
