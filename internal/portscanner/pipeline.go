package portscanner

import "time"

// Pipeline wires together scanning, filtering, diffing, rate-limiting, and
// aggregation into a single reusable processing step.
type Pipeline struct {
	scanner     *Scanner
	filter      *Filter
	rateLimiter *RateLimiter
	aggregator  *Aggregator
	history     *History
}

// PipelineConfig holds the tuning knobs for a Pipeline.
type PipelineConfig struct {
	// Cooldown is the minimum time between repeated alerts for the same port.
	Cooldown time.Duration
	// MaxHistory is the number of snapshots retained in memory.
	MaxHistory int
	// FilterOpts are functional options applied to the Filter.
	FilterOpts []FilterOption
}

// NewPipeline constructs a Pipeline with the supplied configuration.
func NewPipeline(scanner *Scanner, cfg PipelineConfig) *Pipeline {
	if cfg.MaxHistory <= 0 {
		cfg.MaxHistory = 10
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = 30 * time.Second
	}
	return &Pipeline{
		scanner:     scanner,
		filter:      NewFilter(cfg.FilterOpts...),
		rateLimiter: NewRateLimiter(cfg.Cooldown),
		aggregator:  NewAggregator(),
		history:     NewHistory(cfg.MaxHistory),
	}
}

// Run executes one full scan-diff-filter-ratelimit-aggregate cycle and returns
// the resulting ChangeEvents. It returns nil (not an error) when there are no
// noteworthy changes.
func (p *Pipeline) Run(now time.Time) ([]ChangeEvent, error) {
	entries, err := p.scanner.Scan()
	if err != nil {
		return nil, err
	}

	filtered := p.filter.Apply(entries)
	snap := NewSnapshot(filtered, now)
	p.history.Add(snap)

	prev := p.history.Previous()
	if prev == nil {
		// First scan — no previous state to diff against.
		return nil, nil
	}

	events := Diff(prev.ToMap(), snap.ToMap())
	var allowed []ChangeEvent
	for _, ev := range events {
		if p.rateLimiter.Allow(ev, now) {
			allowed = append(allowed, ev)
		}
	}

	return p.aggregator.Aggregate(allowed, now), nil
}

// History returns the underlying History store for inspection.
func (p *Pipeline) History() *History { return p.history }
