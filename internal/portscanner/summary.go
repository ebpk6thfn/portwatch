package portscanner

import (
	"fmt"
	"strings"
	"time"
)

// SummaryReport holds a periodic summary of port activity.
type SummaryReport struct {
	Period    time.Duration
	GeneratedAt time.Time
	Opened    int
	Closed    int
	Suppressed int
	TopPorts  []string
	Anomalies int
}

// SummaryBuilder accumulates ChangeEvents over a window and produces a SummaryReport.
type SummaryBuilder struct {
	period     time.Duration
	events     []ChangeEvent
	suppressed int
	anomalies  int
	now        func() time.Time
}

// NewSummaryBuilder creates a SummaryBuilder for the given reporting period.
func NewSummaryBuilder(period time.Duration, now func() time.Time) *SummaryBuilder {
	if now == nil {
		now = time.Now
	}
	return &SummaryBuilder{period: period, now: now}
}

// Record adds a ChangeEvent to the builder.
func (s *SummaryBuilder) Record(e ChangeEvent) {
	s.events = append(s.events, e)
}

// RecordSuppressed increments the suppressed counter.
func (s *SummaryBuilder) RecordSuppressed() { s.suppressed++ }

// RecordAnomaly increments the anomaly counter.
func (s *SummaryBuilder) RecordAnomaly() { s.anomalies++ }

// Build produces a SummaryReport and resets internal state.
func (s *SummaryBuilder) Build() SummaryReport {
	opened, closed := 0, 0
	portCount := map[string]int{}
	for _, e := range s.events {
		if e.Type == EventOpened {
			opened++
		} else {
			closed++
		}
		key := fmt.Sprintf("%s/%d", e.Entry.Protocol, e.Entry.Port)
		portCount[key]++
	}
	top := topN(portCount, 5)
	r := SummaryReport{
		Period:      s.period,
		GeneratedAt: s.now(),
		Opened:      opened,
		Closed:      closed,
		Suppressed:  s.suppressed,
		TopPorts:    top,
		Anomalies:   s.anomalies,
	}
	s.events = s.events[:0]
	s.suppressed = 0
	s.anomalies = 0
	return r
}

func topN(m map[string]int, n int) []string {
	type kv struct{ k string; v int }
	var pairs []kv
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].v > pairs[i].v {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	var out []string
	for i := 0; i < len(pairs) && i < n; i++ {
		out = append(out, pairs[i].k)
	}
	return out
}

// String returns a human-readable summary.
func (r SummaryReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[Summary %s] opened=%d closed=%d suppressed=%d anomalies=%d",
		r.Period, r.Opened, r.Closed, r.Suppressed, r.Anomalies)
	if len(r.TopPorts) > 0 {
		fmt.Fprintf(&b, " top=[%s]", strings.Join(r.TopPorts, ","))
	}
	return b.String()
}
