package portscanner

import (
	"testing"
	"time"
)

func fixedTokenNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestTokenBucket_FullOnCreate(t *testing.T) {
	tb := NewTokenBucket(TokenBucketPolicy{Max: 5, Rate: 1})
	if got := tb.Tokens(); got != 5 {
		t.Fatalf("expected 5 tokens, got %v", got)
	}
}

func TestTokenBucket_Allow_ConsumesToken(t *testing.T) {
	tb := NewTokenBucket(TokenBucketPolicy{Max: 3, Rate: 1})
	if !tb.Allow() {
		t.Fatal("expected first Allow to succeed")
	}
	if tb.Tokens() >= 3 {
		t.Fatal("expected tokens to decrease after Allow")
	}
}

func TestTokenBucket_ExhaustsAtMax(t *testing.T) {
	base := time.Unix(1000, 0)
	tb := NewTokenBucket(TokenBucketPolicy{Max: 3, Rate: 0.001})
	tb.nowFn = fixedTokenNow(base)
	tb.lastFill = base
	tb.tokens = 3

	for i := 0; i < 3; i++ {
		if !tb.Allow() {
			t.Fatalf("expected Allow to succeed on attempt %d", i)
		}
	}
	if tb.Allow() {
		t.Fatal("expected Allow to fail after exhaustion")
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	base := time.Unix(1000, 0)
	tb := NewTokenBucket(TokenBucketPolicy{Max: 5, Rate: 2})
	tb.nowFn = fixedTokenNow(base)
	tb.lastFill = base
	tb.tokens = 0

	// advance 1 second => +2 tokens
	tb.nowFn = fixedTokenNow(base.Add(time.Second))
	tokens := tb.Tokens()
	if tokens < 1.9 || tokens > 2.1 {
		t.Fatalf("expected ~2 tokens after 1s refill, got %v", tokens)
	}
}

func TestTokenBucket_RefillCappedAtMax(t *testing.T) {
	base := time.Unix(1000, 0)
	tb := NewTokenBucket(TokenBucketPolicy{Max: 5, Rate: 10})
	tb.nowFn = fixedTokenNow(base)
	tb.lastFill = base
	tb.tokens = 0

	// advance 10 seconds => would be 100 tokens but capped at 5
	tb.nowFn = fixedTokenNow(base.Add(10 * time.Second))
	if got := tb.Tokens(); got != 5 {
		t.Fatalf("expected tokens capped at max=5, got %v", got)
	}
}

func TestTokenBucket_DefaultPolicy_UsedOnZeroValues(t *testing.T) {
	tb := NewTokenBucket(TokenBucketPolicy{})
	def := DefaultTokenBucketPolicy()
	if tb.max != def.Max {
		t.Fatalf("expected max=%v, got %v", def.Max, tb.max)
	}
	if tb.rate != def.Rate {
		t.Fatalf("expected rate=%v, got %v", def.Rate, tb.rate)
	}
}
