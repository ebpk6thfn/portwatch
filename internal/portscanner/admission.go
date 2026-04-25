package portscanner

import (
	"fmt"
	"sync"
	"time"
)

// AdmissionPolicy controls the rules used by the Admission controller.
type AdmissionPolicy struct {
	// MaxQueueDepth is the maximum number of events allowed to queue before
	// new arrivals are rejected outright.
	MaxQueueDepth int

	// MinSeverity is the lowest severity level that will be admitted.
	// Events below this level are dropped at the gate.
	MinSeverity Severity

	// CooldownPeriod is the minimum time between admitting events with the
	// same key. Zero means no cooldown is enforced.
	CooldownPeriod time.Duration
}

// DefaultAdmissionPolicy returns a conservative admission policy suitable
// for most deployments.
func DefaultAdmissionPolicy() AdmissionPolicy {
	return AdmissionPolicy{
		MaxQueueDepth:  512,
		MinSeverity:    SeverityLow,
		CooldownPeriod: 0,
	}
}

// AdmissionResult describes why an event was accepted or rejected.
type AdmissionResult int

const (
	AdmissionAllowed  AdmissionResult = iota
	AdmissionRejectedDepth            // queue is full
	AdmissionRejectedSeverity         // event severity too low
	AdmissionRejectedCooldown         // key is in cooldown
)

func (r AdmissionResult) String() string {
	switch r {
	case AdmissionAllowed:
		return "allowed"
	case AdmissionRejectedDepth:
		return "rejected:depth"
	case AdmissionRejectedSeverity:
		return "rejected:severity"
	case AdmissionRejectedCooldown:
		return "rejected:cooldown"
	default:
		return fmt.Sprintf("unknown(%d)", int(r))
	}
}

// Admission is a gate that decides whether a ChangeEvent may proceed into
// the downstream pipeline. It enforces queue-depth limits, minimum severity
// requirements, and optional per-key cooldowns.
type Admission struct {
	mu       sync.Mutex
	policy   AdmissionPolicy
	depth    int
	cooldown map[string]time.Time
	now      func() time.Time
}

// NewAdmission creates an Admission gate with the given policy.
func NewAdmission(policy AdmissionPolicy) *Admission {
	return &Admission{
		policy:   policy,
		cooldown: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Admit evaluates whether the event should be allowed through the gate.
// The caller must call Release() after the event has been processed so that
// the internal depth counter is decremented correctly.
func (a *Admission) Admit(event ChangeEvent) AdmissionResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Severity gate.
	if event.Severity < a.policy.MinSeverity {
		return AdmissionRejectedSeverity
	}

	// Queue-depth gate.
	if a.policy.MaxQueueDepth > 0 && a.depth >= a.policy.MaxQueueDepth {
		return AdmissionRejectedDepth
	}

	// Cooldown gate.
	if a.policy.CooldownPeriod > 0 {
		key := event.Entry.Key()
		if until, ok := a.cooldown[key]; ok && a.now().Before(until) {
			return AdmissionRejectedCooldown
		}
		a.cooldown[key] = a.now().Add(a.policy.CooldownPeriod)
	}

	a.depth++
	return AdmissionAllowed
}

// Release decrements the internal depth counter. It must be called once for
// every event that received AdmissionAllowed from Admit.
func (a *Admission) Release() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.depth > 0 {
		a.depth--
	}
}

// Depth returns the current number of in-flight admitted events.
func (a *Admission) Depth() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.depth
}

// FlushCooldowns removes expired cooldown entries to prevent unbounded growth.
func (a *Admission) FlushCooldowns() {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	for k, until := range a.cooldown {
		if now.After(until) {
			delete(a.cooldown, k)
		}
	}
}
