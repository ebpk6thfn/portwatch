package portscanner

import "time"

// RateLimitPolicy defines rules for rate limiting events by severity or protocol.
type RateLimitPolicy struct {
	DefaultCooldown  time.Duration
	HighCooldown     time.Duration
	MediumCooldown   time.Duration
	LowCooldown      time.Duration
	ProtocolOverride map[string]time.Duration
}

// DefaultRateLimitPolicy returns a sensible default policy.
func DefaultRateLimitPolicy() RateLimitPolicy {
	return RateLimitPolicy{
		DefaultCooldown: 30 * time.Second,
		HighCooldown:    5 * time.Second,
		MediumCooldown:  15 * time.Second,
		LowCooldown:     60 * time.Second,
		ProtocolOverride: map[string]time.Duration{},
	}
}

// CooldownFor returns the cooldown duration for a given severity and protocol.
func (p RateLimitPolicy) CooldownFor(severity, protocol string) time.Duration {
	if d, ok := p.ProtocolOverride[protocol]; ok {
		return d
	}
	switch severity {
	case "high":
		return p.HighCooldown
	case "medium":
		return p.MediumCooldown
	case "low":
		return p.LowCooldown
	default:
		return p.DefaultCooldown
	}
}

// PolicyRateLimiter applies a RateLimitPolicy to filter ChangeEvents.
type PolicyRateLimiter struct {
	policy  RateLimitPolicy
	cooldown *Cooldown
}

// NewPolicyRateLimiter creates a PolicyRateLimiter using the given policy.
func NewPolicyRateLimiter(policy RateLimitPolicy, now func() interface{}) *PolicyRateLimiter {
	return &PolicyRateLimiter{
		policy:  policy,
		cooldown: NewCooldown(policy.DefaultCooldown),
	}
}

// Allow returns true if the event should be forwarded given the policy cooldown.
func (r *PolicyRateLimiter) Allow(event ChangeEvent) bool {
	key := dedupKey(event)
	cooldown := r.policy.CooldownFor(string(event.Severity), event.Entry.Protocol)
	r.cooldown.SetPeriod(cooldown)
	return r.cooldown.Allow(key)
}
