package portscanner

import "time"

// BurstAlert wraps a BurstDetector and produces a ChangeEvent summary
// when a burst is detected.
type BurstAlert struct {
	detector  *BurstDetector
	protocol  string
	cooldown  *Cooldown
}

// NewBurstAlert creates a BurstAlert for the given protocol.
// Cooldown prevents repeated burst alerts within a short period.
func NewBurstAlert(threshold int, window, alertCooldown time.Duration, protocol string) *BurstAlert {
	return &BurstAlert{
		detector: NewBurstDetector(threshold, window),
		protocol: protocol,
		cooldown: NewCooldown(alertCooldown),
	}
}

// Observe records an event and returns a synthetic ChangeEvent if a burst
// is detected and the alert cooldown has elapsed; otherwise returns nil.
func (ba *BurstAlert) Observe(e ChangeEvent) *ChangeEvent {
	if !ba.detector.Record() {
		return nil
	}
	key := "burst:" + ba.protocol
	if !ba.cooldown.Allow(key) {
		return nil
	}
	synthetic := ChangeEvent{
		Entry: Entry{
			Protocol: ba.protocol,
			Port:     0,
			Process:  "[burst-alert]",
		},
		Type:     EventOpened,
		Severity: SeverityHigh,
		Label:    "burst-detected",
	}
	return &synthetic
}

// Reset clears the burst detector's event history and resets the cooldown
// state for this protocol. Useful for testing or after a known noisy period.
func (ba *BurstAlert) Reset() {
	ba.detector.Reset()
	ba.cooldown.Reset("burst:" + ba.protocol)
}
