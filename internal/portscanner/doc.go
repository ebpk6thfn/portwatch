// Package portscanner provides utilities for scanning active network ports
// on a Linux host by reading /proc/net/tcp and /proc/net/tcp6, diffing
// snapshots to detect changes, and resolving socket inodes to owning processes.
//
// # Core types
//
//   - Scanner: reads /proc/net/tcp[6] and returns a slice of Entry values.
//   - Entry: represents a single listening port (protocol, address, port, process).
//   - Snapshot: an immutable, timestamped view of all active ports.
//   - Diff: compares two Snapshots and returns opened/closed ChangeEvents.
//   - Filter: applies user-defined rules (exclude ports, loopback, private ranges).
//   - StateStore: persists the last snapshot to disk for cross-run diffing.
//   - History: keeps a bounded ring of recent snapshots for trend analysis.
//   - RateLimiter: suppresses duplicate events within a configurable cooldown.
//   - Aggregator: batches ChangeEvents from a scan cycle into a single report.
//   - Resolver: maps socket inodes to PID and process name via /proc/<pid>/fd.
package portscanner
