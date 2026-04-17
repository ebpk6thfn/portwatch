package portscanner

import (
	"fmt"
	"time"
)

// AnomalyType classifies the kind of anomaly detected.
type AnomalyType string

const (
	AnomalyBurst    AnomalyType = "burst"
	AnomalyRapidCycle AnomalyType = "rapid_cycle"
	AnomalyUnknown  AnomalyType = "unknown"
)

// Anomaly represents a detected anomalous pattern in port activity.
type Anomaly struct {
	Type      AnomalyType
	Port      uint16
	Protocol  string
	Score     float64
	DetectedAt time.Time
	Detail    string
}

func (a Anomaly) String() string {
	return fmt.Sprintf("[%s] %s/%d score=%.2f at %s: %s",
		a.Type, a.Protocol, a.Port, a.Score,
		a.DetectedAt.Format(time.RFC3339), a.Detail)
}

// AnomalyDetector combines burst and rapid-cycle detection into one pass.
type AnomalyDetector struct {
	burst  *BurstDetector
	decay  *DecayCounter
	threshold int
	cooldown  *Cooldown
}

// NewAnomalyDetector creates a detector with the given burst window/threshold
// and a decay half-life for rapid-cycle scoring.
func NewAnomalyDetector(window time.Duration, burstThreshold int, halfLife time.Duration, cooldownPeriod time.Duration) *AnomalyDetector {
	return &AnomalyDetector{
		burst:     NewBurstDetector(window, burstThreshold),
		decay:     NewDecayCounter(halfLife),
		threshold: burstThreshold,
		cooldown:  NewCooldown(cooldownPeriod),
	}
}

// Evaluate checks a ChangeEvent for anomalies and returns any detected, or nil.
func (d *AnomalyDetector) Evaluate(ev ChangeEvent, now time.Time) *Anomaly {
	key := ev.Entry.Key()

	isBurst, count := d.burst.Record(ev, now)
	decayScore := d.decay.Add(key, now)

	var aType AnomalyType
	var detail string
	var score float64

	switch {
	case isBurst:
		aType = AnomalyBurst
		score = float64(count) / float64(d.threshold)
		detail = fmt.Sprintf("%d events in window", count)
	case decayScore > 1.5:
		aType = AnomalyRapidCycle
		score = decayScore
		detail = fmt.Sprintf("decay score %.2f", decayScore)
	default:
		return nil
	}

	if !d.cooldown.Allow(key, now) {
		return nil
	}

	return &Anomaly{
		Type:       aType,
		Port:       ev.Entry.Port,
		Protocol:   ev.Entry.Protocol,
		Score:      score,
		DetectedAt: now,
		Detail:     detail,
	}
}
