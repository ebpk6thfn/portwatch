package portscanner

import (
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter that refills at a fixed rate.
// It is safe for concurrent use.
type TokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastFill time.Time
	nowFn    func() time.Time
}

// TokenBucketPolicy holds configuration for a TokenBucket.
type TokenBucketPolicy struct {
	// Max is the maximum number of tokens (burst capacity).
	Max float64
	// Rate is the number of tokens added per second.
	Rate float64
}

// DefaultTokenBucketPolicy returns a sensible default policy.
func DefaultTokenBucketPolicy() TokenBucketPolicy {
	return TokenBucketPolicy{
		Max:  10,
		Rate: 2,
	}
}

// NewTokenBucket creates a new TokenBucket filled to capacity.
func NewTokenBucket(p TokenBucketPolicy) *TokenBucket {
	if p.Max <= 0 {
		p.Max = DefaultTokenBucketPolicy().Max
	}
	if p.Rate <= 0 {
		p.Rate = DefaultTokenBucketPolicy().Rate
	}
	return &TokenBucket{
		tokens:   p.Max,
		max:      p.Max,
		rate:     p.Rate,
		lastFill: time.Now(),
		nowFn:    time.Now,
	}
}

// Allow attempts to consume one token. Returns true if a token was available.
func (t *TokenBucket) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.refill()
	if t.tokens >= 1 {
		t.tokens--
		return true
	}
	return false
}

// Tokens returns the current number of available tokens.
func (t *TokenBucket) Tokens() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.refill()
	return t.tokens
}

// refill adds tokens based on elapsed time since last fill. Must be called with lock held.
func (t *TokenBucket) refill() {
	now := t.nowFn()
	elapsed := now.Sub(t.lastFill).Seconds()
	if elapsed > 0 {
		t.tokens = min64(t.tokens+elapsed*t.rate, t.max)
		t.lastFill = now
	}
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
