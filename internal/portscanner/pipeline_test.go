package portscanner

import (
	"testing"
	"time"
)

// stubScanner satisfies the same interface used by Pipeline but returns
// a pre-configured slice of entries without touching /proc.
type stubScanner struct {
	entries []Entry
	err     error
}

func (s *stubScanner) Scan() ([]Entry, error) { return s.entries, s.err }

// newPipelineForTest builds a Pipeline backed by a stubScanner.
func newPipelineForTest(entries []Entry, cfg PipelineConfig) (*Pipeline, *stubScanner) {
	stub := &stubScanner{entries: entries}
	// We cannot pass *stubScanner where *Scanner is expected directly, so we
	// exercise the public constructor with a real (zero-value) Scanner and
	// override the internal scanner field via the struct literal to keep the
	// test self-contained.
	p := NewPipeline(&Scanner{}, cfg)
	p.scanner = &Scanner{} // reset; we'll swap the scan func via monkey-free approach
	// Instead, build the pipeline manually to avoid /proc dependency.
	p2 := &Pipeline{
		scanner:     &Scanner{},
		filter:      NewFilter(cfg.FilterOpts...),
		rateLimiter: NewRateLimiter(cfg.Cooldown),
		aggregator:  NewAggregator(),
		history:     NewHistory(cfg.MaxHistory),
	}
	_ = p
	return p2, stub
}

func TestPipeline_FirstRunNoEvents(t *testing.T) {
	cfg := PipelineConfig{Cooldown: time.Second, MaxHistory: 5}
	pipe, _ := newPipelineForTest(nil, cfg)

	// Manually inject a snapshot as the first run (simulates scanner output).
	now := time.Now()
	pipe.history.Add(NewSnapshot([]Entry{makePipeEntry("127.0.0.1", 8080, "tcp")}, now))

	// Only one snapshot — Previous() is nil, so Run equivalent yields nothing.
	if pipe.history.Len() != 1 {
		t.Fatalf("expected 1 snapshot, got %d", pipe.history.Len())
	}
	if pipe.history.Previous() != nil {
		t.Fatal("expected nil previous on first snapshot")
	}
}

func TestPipeline_DetectsOpenedPort(t *testing.T) {
	now := time.Now()
	cfg := PipelineConfig{Cooldown: time.Second, MaxHistory: 5}
	pipe, _ := newPipelineForTest(nil, cfg)

	e1 := makePipeEntry("0.0.0.0", 9090, "tcp")
	e2 := makePipeEntry("0.0.0.0", 9091, "tcp")

	pipe.history.Add(NewSnapshot([]Entry{e1}, now))
	pipe.history.Add(NewSnapshot([]Entry{e1, e2}, now.Add(time.Second)))

	events := Diff(pipe.history.Previous().ToMap(), pipe.history.Latest().ToMap())
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventOpened {
		t.Errorf("expected Opened, got %v", events[0].Type)
	}
}

func TestPipeline_RateLimiterSuppressesDuplicate(t *testing.T) {
	now := time.Now()
	cfg := PipelineConfig{Cooldown: time.Minute, MaxHistory: 5}
	pipe, _ := newPipelineForTest(nil, cfg)

	e := makePipeEntry("0.0.0.0", 7070, "tcp")
	ev := ChangeEvent{Type: EventOpened, Entry: e, At: now}

	if !pipe.rateLimiter.Allow(ev, now) {
		t.Fatal("first event should be allowed")
	}
	if pipe.rateLimiter.Allow(ev, now.Add(time.Second)) {
		t.Fatal("duplicate within cooldown should be suppressed")
	}
}

func makePipeEntry(ip string, port uint16, proto string) Entry {
	return Entry{LocalIP: ip, LocalPort: port, Protocol: proto}
}
