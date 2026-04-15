package notifier

// Event represents a port change event to be dispatched.
type Event struct {
	Type    string `json:"type"`    // "opened" or "closed"
	Port    uint16 `json:"port"`
	Proto   string `json:"proto"`
	PID     int    `json:"pid"`
	Process string `json:"process"`
}

// Notifier is the interface implemented by all notification backends.
type Notifier interface {
	Notify(event Event) error
	Name() string
}

// Multi dispatches events to multiple notifiers in order.
type Multi struct {
	notifiers []Notifier
}

// NewMulti creates a Multi notifier wrapping the provided backends.
func NewMulti(notifiers ...Notifier) *Multi {
	return &Multi{notifiers: notifiers}
}

// Notify sends the event to all registered notifiers, collecting errors.
func (m *Multi) Notify(event Event) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Notify(event); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", n.Name(), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notifier errors: %v", errs)
	}
	return nil
}

func (m *Multi) Name() string { return "multi" }
