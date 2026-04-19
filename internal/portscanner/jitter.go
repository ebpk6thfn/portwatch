package portscanner

import (
	"math/rand"
	"time"
)

// Jitter adds randomized delay variation to a base duration to prevent
// thundering-herd effects when multiple scanners fire simultaneously.
type Jitter struct {
	rng      *rand.Rand
	factor   float64 // fraction of base to vary, e.g. 0.2 = ±20%
}

// NewJitter returns a Jitter with the given spread factor (0.0–1.0).
// A factor of 0.2 means the returned duration will be base ± 20% of base.
func NewJitter(factor float64) *Jitter {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return &Jitter{
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		factor: factor,
	}
}

// Apply returns base adjusted by a random value within ±factor*base.
// The result is always >= 0.
func (j *Jitter) Apply(base time.Duration) time.Duration {
	if base <= 0 || j.factor == 0 {
		return base
	}
	spread := float64(base) * j.factor
	// random value in [-spread, +spread]
	delta := (j.rng.Float64()*2 - 1) * spread
	result := time.Duration(float64(base) + delta)
	if result < 0 {
		return 0
	}
	return result
}

// ApplyPositive returns base plus a random value in [0, factor*base],
// i.e. only increases the duration. Useful for retry back-off spreading.
func (j *Jitter) ApplyPositive(base time.Duration) time.Duration {
	if base <= 0 || j.factor == 0 {
		return base
	}
	spread := float64(base) * j.factor
	delta := j.rng.Float64() * spread
	return base + time.Duration(delta)
}
