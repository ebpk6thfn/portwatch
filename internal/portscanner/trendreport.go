package portscanner

import (
	"fmt"
	"strings"
	"time"
)

// TrendReport summarises the current state of a TrendTracker.
type TrendReport struct {
	Direction  TrendDirection
	PointCount int
	First      int
	Last       int
	Window     time.Duration
}

// String returns a human-readable one-line summary.
func (r TrendReport) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "trend=%s window=%s points=%d",
		r.Direction, r.Window.Round(time.Second), r.PointCount)
	if r.PointCount >= 2 {
		fmt.Fprintf(&sb, " first=%d last=%d", r.First, r.Last)
	}
	return sb.String()
}

// BuildReport generates a TrendReport from the tracker at the given time.
func BuildReport(tr *TrendTracker, now time.Time) TrendReport {
	pts := tr.Points(now)
	rep := TrendReport{
		Direction:  tr.Trend(now),
		PointCount: len(pts),
		Window:     tr.window,
	}
	if len(pts) >= 1 {
		rep.First = pts[0].Count
		rep.Last = pts[len(pts)-1].Count
	}
	return rep
}
