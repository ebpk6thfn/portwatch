package notifier

import (
	"fmt"
	"log"
)

// RoutedDispatcher holds a map of label -> Notifier and dispatches
// pre-routed event payloads to the appropriate notifier.
type RoutedDispatcher struct {
	sinks map[string]Notifier
	fallback Notifier
}

// NewRoutedDispatcher creates a dispatcher. fallback is used when no sink
// matches a label; it may be nil to silently drop unmatched events.
func NewRoutedDispatcher(sinks map[string]Notifier, fallback Notifier) *RoutedDispatcher {
	return &RoutedDispatcher{sinks: sinks, fallback: fallback}
}

// Dispatch sends msg to the notifier registered under label.
func (d *RoutedDispatcher) Dispatch(label, msg string) error {
	n, ok := d.sinks[label]
	if !ok {
		if d.fallback != nil {
			return d.fallback.Notify(Event{Message: msg})
		}
		log.Printf("router_dispatch: no sink for label %q, dropping", label)
		return nil
	}
	return n.Notify(Event{Message: msg})
}

// DispatchAll sends each label's messages in batch.
func (d *RoutedDispatcher) DispatchAll(groups map[string][]string) error {
	var first error
	for label, msgs := range groups {
		for _, msg := range msgs {
			if err := d.Dispatch(label, msg); err != nil && first == nil {
				first = fmt.Errorf("dispatch %s: %w", label, err)
			}
		}
	}
	return first
}
