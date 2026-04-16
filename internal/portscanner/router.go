package portscanner

// RouteRule defines a condition and destination label for routing events.
type RouteRule struct {
	Severity  string // if non-empty, match on severity
	Protocol  string // if non-empty, match on protocol
	DestLabel string // label to attach when rule matches
}

// Router routes ChangeEvents to named destinations based on rules.
type Router struct {
	rules []RouteRule
	defaultLabel string
}

// NewRouter creates a Router with the given rules and a fallback default label.
func NewRouter(defaultLabel string, rules []RouteRule) *Router {
	return &Router{rules: rules, defaultLabel: defaultLabel}
}

// Route evaluates rules against the event and returns the destination label.
func (r *Router) Route(e ChangeEvent) string {
	for _, rule := range r.rules {
		if rule.Severity != "" && e.Severity != rule.Severity {
			continue
		}
		if rule.Protocol != "" && e.Entry.Protocol != rule.Protocol {
			continue
		}
		return rule.DestLabel
	}
	return r.defaultLabel
}

// RouteAll applies Route to every event and returns a map of label -> events.
func (r *Router) RouteAll(events []ChangeEvent) map[string][]ChangeEvent {
	out := make(map[string][]ChangeEvent)
	for _, e := range events {
		label := r.Route(e)
		out[label] = append(out[label], e)
	}
	return out
}
