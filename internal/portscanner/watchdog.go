package portscanner

import (
	"sync"
	"time"
)

// WatchdogPolicy defines the configuration for the watchdog.
type WatchdogPolicy struct {
	// MaxSilence is the maximum duration without a successful scan before
	// the watchdog fires.
	MaxSilence time.Duration
	// CheckInterval is how often the watchdog inspects the last-seen timestamp.
	CheckInterval time.Duration
}

// DefaultWatchdogPolicy returns a sensible default watchdog policy.
func DefaultWatchdogPolicy() WatchdogPolicy {
	return WatchdogPolicy{
		MaxSilence:    2 * time.Minute,
		CheckInterval: 15 * time.Second,
	}
}

// Watchdog monitors the liveness of the scan pipeline. If no scan heartbeat
// is recorded within MaxSilence, the provided alert function is called.
type Watchdog struct {
	mu       sync.Mutex
	policy   WatchdogPolicy
	lastBeat time.Time
	stop     chan struct{}
	alert    func(staleDuration time.Duration)
	now      func() time.Time
}

// NewWatchdog creates a Watchdog with the given policy and alert callback.
func NewWatchdog(policy WatchdogPolicy, alert func(staleDuration time.Duration)) *Watchdog {
	return &Watchdog{
		policy: policy,
		alert:  alert,
		stop:   make(chan struct{}),
		now:    time.Now,
	}
}

// Beat records a successful scan heartbeat.
func (w *Watchdog) Beat() {
	w.mu.Lock()
	w.lastBeat = w.now()
	w.mu.Unlock()
}

// Start begins the background liveness check loop.
func (w *Watchdog) Start() {
	w.mu.Lock()
	w.lastBeat = w.now()
	w.mu.Unlock()

	go func() {
		ticker := time.NewTicker(w.policy.CheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-w.stop:
				return
			case <-ticker.C:
				w.mu.Lock()
				staleDur := w.now().Sub(w.lastBeat)
				w.mu.Unlock()
				if staleDur > w.policy.MaxSilence {
					w.alert(staleDur)
				}
			}
		}
	}()
}

// Stop shuts down the watchdog loop.
func (w *Watchdog) Stop() {
	close(w.stop)
}

// StaleDuration returns how long it has been since the last heartbeat.
func (w *Watchdog) StaleDuration() time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.now().Sub(w.lastBeat)
}
