package portscanner

import (
	"net"
	"testing"
	"time"
)

func makeRLEntry(port uint16, proto string) Entry {
	return Entry{
		LocalAddr: net.ParseIP("127.0.0.1"),
		LocalPort: port,
		Protocol:  proto,
		State:     "LISTEN",
	}
}

func TestRateLimiter_AllowsFirst(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)
	if !rl.Allow("tcp:8080") {
		t.Fatal("expected first event to be allowed")
	}
}

func TestRateLimiter_SuppressesWithinCooldown(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(10 * time.Second)
	rl.now = func() time.Time { return base }

	rl.Allow("tcp:8080")

	// Still within cooldown window
	rl.now = func() time.Time { return base.Add(3 * time.Second) }
	if rl.Allow("tcp:8080") {
		t.Fatal("expected event to be suppressed within cooldown")
	}
}

func TestRateLimiter_AllowsAfterCooldown(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(5 * time.Second)
	rl.now = func() time.Time { return base }

	rl.Allow("tcp:9090")

	// Past cooldown window
	rl.now = func() time.Time { return base.Add(6 * time.Second) }
	if !rl.Allow("tcp:9090") {
		t.Fatal("expected event to be allowed after cooldown expires")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(10 * time.Second)
	rl.Allow("tcp:8080")

	if !rl.Allow("tcp:9090") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestRateLimiter_Filter(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(10 * time.Second)
	rl.now = func() time.Time { return base }

	events := []ChangeEvent{
		{Entry: makeRLEntry(8080, "tcp"), Kind: EventOpened},
		{Entry: makeRLEntry(8080, "tcp"), Kind: EventOpened}, // duplicate
		{Entry: makeRLEntry(9090, "tcp"), Kind: EventOpened},
	}

	result := rl.Filter(events)
	if len(result) != 2 {
		t.Fatalf("expected 2 filtered events, got %d", len(result))
	}
}

func TestRateLimiter_Purge(t *testing.T) {
	base := time.Now()
	rl := NewRateLimiter(5 * time.Second)
	rl.now = func() time.Time { return base }

	rl.Allow("tcp:1234")
	rl.Allow("tcp:5678")

	// Advance past cooldown
	rl.now = func() time.Time { return base.Add(6 * time.Second) }
	rl.Purge()

	rl.mu.Lock()
	n := len(rl.seen)
	rl.mu.Unlock()

	if n != 0 {
		t.Fatalf("expected seen map to be empty after purge, got %d entries", n)
	}
}
