package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeEvent(t string, port uint16) Event {
	return Event{Type: t, Port: port, Proto: "tcp", PID: 1234, Process: "nginx"}
}

// TestWebhookNotifier_Success verifies a 200 response causes no error.
func TestWebhookNotifier_Success(t *testing.T) {
	var received Event
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	w := NewWebhook(ts.URL)
	ev := makeEvent("opened", 8080)
	if err := w.Notify(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Port != 8080 {
		t.Errorf("expected port 8080, got %d", received.Port)
	}
}

// TestWebhookNotifier_Non2xx verifies a non-2xx response returns an error.
func TestWebhookNotifier_Non2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	w := NewWebhook(ts.URL)
	if err := w.Notify(makeEvent("closed", 443)); err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// TestMultiNotifier_AllCalled verifies all backends receive the event.
func TestMultiNotifier_AllCalled(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	w1 := NewWebhook(ts.URL)
	w2 := NewWebhook(ts.URL)
	m := NewMulti(w1, w2)

	if err := m.Notify(makeEvent("opened", 3000)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 webhook calls, got %d", calls)
	}
}
