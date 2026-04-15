package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier sends port change events as JSON POST requests.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// NewWebhook creates a WebhookNotifier targeting the given URL.
func NewWebhook(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Notify serialises the event and POSTs it to the configured URL.
func (w *WebhookNotifier) Notify(event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("webhook marshal: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook non-2xx response: %d", resp.StatusCode)
	}
	return nil
}

func (w *WebhookNotifier) Name() string { return "webhook" }
