// Package portscanner provides port scanning, diffing, filtering, and
// event pipeline primitives for portwatch.
//
// StateChangeTracker
//
// StateChangeTracker records the first and last observed time for any
// string key, typically a port Entry key such as "tcp:80". It is used
// in the event pipeline to annotate alerts with how long a port has
// been in its current state, enabling duration-aware suppression and
// alerting rules.
//
// Typical usage:
//
//	tracker := portscanner.NewStateChangeTracker(nil)
//
//	// On port open:
//	tracker.Record(entry.Key())
//
//	// On port close — check how long it was open:
//	_, duration := tracker.Record(entry.Key())
//	tracker.Forget(entry.Key())
package portscanner
