package portscanner

import (
	"testing"
	"time"
)

func makeHeatmapEvent(port uint16, proto string) ChangeEvent {
	return ChangeEvent{
		Entry: Entry{
			Port:     port,
			Protocol: proto,
		},
		Kind: EventOpened,
	}
}

func TestHeatmap_Empty_ReturnsNone(t *testing.T) {
	h := NewHeatmap(time.Minute)
	now := time.Now()
	top := h.Top(10, now)
	if len(top) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(top))
	}
}

func TestHeatmap_RecordAndTop(t *testing.T) {
	h := NewHeatmap(time.Minute)
	now := time.Now()

	h.Record(makeHeatmapEvent(80, "tcp"), now)
	h.Record(makeHeatmapEvent(80, "tcp"), now)
	h.Record(makeHeatmapEvent(443, "tcp"), now)

	top := h.Top(10, now)
	if len(top) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(top))
	}
	if top[0].Port != 80 || top[0].Hits != 2 {
		t.Errorf("expected port 80 with 2 hits, got port=%d hits=%d", top[0].Port, top[0].Hits)
	}
	if top[1].Port != 443 || top[1].Hits != 1 {
		t.Errorf("expected port 443 with 1 hit, got port=%d hits=%d", top[1].Port, top[1].Hits)
	}
}

func TestHeatmap_TopLimitedToN(t *testing.T) {
	h := NewHeatmap(time.Minute)
	now := time.Now()

	ports := []uint16{80, 443, 8080, 22, 3306}
	for _, p := range ports {
		h.Record(makeHeatmapEvent(p, "tcp"), now)
	}

	top := h.Top(3, now)
	if len(top) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(top))
	}
}

func TestHeatmap_EvictsExpiredEntries(t *testing.T) {
	h := NewHeatmap(time.Minute)
	past := time.Now().Add(-2 * time.Minute)
	now := time.Now()

	h.Record(makeHeatmapEvent(80, "tcp"), past)
	h.Record(makeHeatmapEvent(443, "tcp"), now)

	top := h.Top(10, now)
	if len(top) != 1 {
		t.Fatalf("expected 1 active entry after eviction, got %d", len(top))
	}
	if top[0].Port != 443 {
		t.Errorf("expected port 443, got %d", top[0].Port)
	}
}

func TestHeatmap_Len(t *testing.T) {
	h := NewHeatmap(time.Minute)
	now := time.Now()

	if h.Len() != 0 {
		t.Fatal("expected empty heatmap")
	}
	h.Record(makeHeatmapEvent(80, "tcp"), now)
	h.Record(makeHeatmapEvent(443, "tcp"), now)
	if h.Len() != 2 {
		t.Fatalf("expected 2, got %d", h.Len())
	}
}

func TestHeatmap_DefaultWindow_WhenZero(t *testing.T) {
	h := NewHeatmap(0)
	if h.window <= 0 {
		t.Error("expected positive default window")
	}
}

func TestHeatmap_HeatmapEntry_String(t *testing.T) {
	e := HeatmapEntry{Port: 80, Protocol: "tcp", Hits: 5, LastSeen: time.Time{}}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string from HeatmapEntry.String()")
	}
}
