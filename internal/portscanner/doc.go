// Package portscanner provides primitives for discovering, tracking, and
// comparing the set of open network ports on the local host.
//
// # Core types
//
//   - [Scanner]      – reads /proc/net/tcp (and tcp6/udp/udp6) to produce
//     a current snapshot of listening ports.
//   - [Entry]        – a single port observation (protocol, address, port,
//     optional process name/PID).
//   - [Snapshot]     – an immutable, indexed collection of [Entry] values
//     captured at a specific point in time.
//   - [Diff]         – compares two snapshots and returns a slice of
//     [ChangeEvent] values describing what opened or closed.
//   - [Filter]       – a composable predicate that can exclude entries by
//     port number, protocol, loopback address, or private subnet.
//   - [RateLimiter]  – suppresses repeated events for the same port within
//     a configurable cooldown window.
//   - [StateStore]   – persists the last-known snapshot to disk so that
//     portwatch can detect changes across process restarts.
//   - [History]      – retains a bounded ring of recent snapshots for
//     in-memory trend analysis.
//   - [Aggregator]   – collects the [ChangeEvent] slice produced by a single
//     scan cycle and merges them into a single [AggregatedEvent] for
//     downstream consumers such as notifiers.
package portscanner
