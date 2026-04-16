package notifier

import (
	"errors"
	"testing"
)

type recordingNotifier struct {
	calls []string
	err   error
}

func (r *recordingNotifier) Notify(e Event) error {
	r.calls = append(r.calls, e.Message)
	return r.err
}

func TestRoutedDispatcher_KnownLabel(t *testing.T) {
	rec := &recordingNotifier{}
	d := NewRoutedDispatcher(map[string]Notifier{"alerts": rec}, nil)
	if err := d.Dispatch("alerts", "hello"); err != nil {
		t.Fatal(err)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "hello" {
		t.Fatalf("unexpected calls: %v", rec.calls)
	}
}

func TestRoutedDispatcher_UnknownLabel_Fallback(t *testing.T) {
	fb := &recordingNotifier{}
	d := NewRoutedDispatcher(nil, fb)
	if err := d.Dispatch("unknown", "msg"); err != nil {
		t.Fatal(err)
	}
	if len(fb.calls) != 1 {
		t.Fatalf("expected fallback called once, got %v", fb.calls)
	}
}

func TestRoutedDispatcher_UnknownLabel_NoFallback_Drops(t *testing.T) {
	d := NewRoutedDispatcher(nil, nil)
	if err := d.Dispatch("ghost", "msg"); err != nil {
		t.Fatalf("expected nil error on drop, got %v", err)
	}
}

func TestRoutedDispatcher_DispatchAll(t *testing.T) {
	a := &recordingNotifier{}
	b := &recordingNotifier{}
	d := NewRoutedDispatcher(map[string]Notifier{"a": a, "b": b}, nil)
	groups := map[string][]string{
		"a": {"a1", "a2"},
		"b": {"b1"},
	}
	if err := d.DispatchAll(groups); err != nil {
		t.Fatal(err)
	}
	if len(a.calls) != 2 {
		t.Fatalf("expected 2 calls to a, got %d", len(a.calls))
	}
	if len(b.calls) != 1 {
		t.Fatalf("expected 1 call to b, got %d", len(b.calls))
	}
}

func TestRoutedDispatcher_DispatchAll_ReturnsFirstError(t *testing.T) {
	bad := &recordingNotifier{err: errors.New("boom")}
	d := NewRoutedDispatcher(map[string]Notifier{"x": bad}, nil)
	err := d.DispatchAll(map[string][]string{"x": {"msg"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
