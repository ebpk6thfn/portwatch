// Package portscanner provides port scanning, diffing, and event processing
// primitives for portwatch.
//
// # RetentionStore
//
// RetentionStore keeps a bounded, time-limited log of ChangeEvents.
// It is useful for surfacing recent activity in status reports or
// diagnostic endpoints without unbounded memory growth.
//
// Events are evicted when they exceed MaxAge (if set) or when the
// total count exceeds MaxCount (if set). Both policies may be combined.
//
// Example:
//
//	store := portscanner.NewRetentionStore(portscanner.RetentionPolicy{
//		MaxAge:   30 * time.Minute,
//		MaxCount: 500,
//	})
//	store.Add(event)
//	recent := store.All()
package portscanner
