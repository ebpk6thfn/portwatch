package portscanner

import (
	"sync"
	"time"
)

// PressureLevel represents the current load level of the event pipeline.
type PressureLevel int

const (
	PressureLow    PressureLevel = iota // below low-water mark
	PressureMedium                      // between low and high water marks
	PressureHigh                        // above high-water mark
)

func (p PressureLevel) String() string {
	switch p {
	case PressureLow:
		return "low"
	case PressureMedium:
		return "medium"
	case PressureHigh:
		return "high"
	default:
		return "unknown"
	}
}

// PressurePolicy configures thresholds for the pressure gauge.
type PressurePolicy struct {
	LowWatermark  int           // depth at or below which pressure is Low
	HighWatermark int           // depth at or above which pressure is High
	Window        time.Duration // window over which depth samples are averaged
}

// DefaultPressurePolicy returns sensible defaults.
func DefaultPressurePolicy() PressurePolicy {
	return PressurePolicy{
		LowWatermark:  10,
		HighWatermark: 50,
		Window:        30 * time.Second,
	}
}

// PressureGauge tracks pipeline depth and reports the current pressure level.
type PressureGauge struct {
	mu      sync.Mutex
	policy  PressurePolicy
	samples []depthSample
	now     func() time.Time
}

type depthSample struct {
	at    time.Time
	depth int
}

// NewPressureGauge creates a PressureGauge with the given policy.
func NewPressureGauge(policy PressurePolicy) *PressureGauge {
	return &PressureGauge{
		policy: policy,
		now:    time.Now,
	}
}

// Record adds a depth observation at the current time.
func (g *PressureGauge) Record(depth int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	g.samples = append(g.samples, depthSample{at: now, depth: depth})
	g.evict(now)
}

// Level returns the current pressure level based on the average depth.
func (g *PressureGauge) Level() PressureLevel {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.evict(g.now())
	if len(g.samples) == 0 {
		return PressureLow
	}
	sum := 0
	for _, s := range g.samples {
		sum += s.depth
	}
	avg := sum / len(g.samples)
	switch {
	case avg >= g.policy.HighWatermark:
		return PressureHigh
	case avg <= g.policy.LowWatermark:
		return PressureLow
	default:
		return PressureMedium
	}
}

// Depth returns the number of retained samples.
func (g *PressureGauge) Depth() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.samples)
}

func (g *PressureGauge) evict(now time.Time) {
	cutoff := now.Add(-g.policy.Window)
	i := 0
	for i < len(g.samples) && g.samples[i].at.Before(cutoff) {
		i++
	}
	g.samples = g.samples[i:]
}
