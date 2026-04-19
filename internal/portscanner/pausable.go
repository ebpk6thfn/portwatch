package portscanner

import "sync"

// Pausable wraps a pipeline stage so it can be temporarily paused.
// While paused, events are dropped rather than forwarded.
type Pausable struct {
	mu     sync.RWMutex
	paused bool
}

// NewPausable returns a new Pausable in the running state.
func NewPausable() *Pausable {
	return &Pausable{}
}

// Pause stops event forwarding.
func (p *Pausable) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = true
}

// Resume restores event forwarding.
func (p *Pausable) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = false
}

// IsPaused returns true if the stage is currently paused.
func (p *Pausable) IsPaused() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.paused
}

// Filter returns only those events that pass through when not paused.
// If paused, an empty slice is returned.
func (p *Pausable) Filter(events []ChangeEvent) []ChangeEvent {
	if p.IsPaused() {
		return nil
	}
	return events
}
