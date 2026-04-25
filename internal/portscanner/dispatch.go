package portscanner

import (
	"sync"
	"time"
)

// DispatchPolicy controls how events are dispatched to handlers.
type DispatchPolicy struct {
	Workers    int
	QueueDepth int
	Timeout    time.Duration
}

// DefaultDispatchPolicy returns a sensible default dispatch policy.
func DefaultDispatchPolicy() DispatchPolicy {
	return DispatchPolicy{
		Workers:    4,
		QueueDepth: 256,
		Timeout:    5 * time.Second,
	}
}

// DispatchHandler is a function that handles a ChangeEvent.
type DispatchHandler func(event ChangeEvent) error

// Dispatcher fans out ChangeEvents to registered handlers concurrently.
type Dispatcher struct {
	policy   DispatchPolicy
	handlers []DispatchHandler
	queue    chan ChangeEvent
	wg       sync.WaitGroup
	once     sync.Once
	stop     chan struct{}
}

// NewDispatcher creates a new Dispatcher with the given policy.
func NewDispatcher(policy DispatchPolicy) *Dispatcher {
	if policy.Workers <= 0 {
		policy.Workers = DefaultDispatchPolicy().Workers
	}
	if policy.QueueDepth <= 0 {
		policy.QueueDepth = DefaultDispatchPolicy().QueueDepth
	}
	if policy.Timeout <= 0 {
		policy.Timeout = DefaultDispatchPolicy().Timeout
	}
	d := &Dispatcher{
		policy: policy,
		queue:  make(chan ChangeEvent, policy.QueueDepth),
		stop:   make(chan struct{}),
	}
	for i := 0; i < policy.Workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
	return d
}

// Register adds a handler to be called for each dispatched event.
func (d *Dispatcher) Register(h DispatchHandler) {
	d.handlers = append(d.handlers, h)
}

// Dispatch enqueues an event for processing. Returns false if the queue is full.
func (d *Dispatcher) Dispatch(event ChangeEvent) bool {
	select {
	case d.queue <- event:
		return true
	default:
		return false
	}
}

// Close shuts down the dispatcher and waits for all workers to finish.
func (d *Dispatcher) Close() {
	d.once.Do(func() {
		close(d.stop)
		close(d.queue)
		d.wg.Wait()
	})
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for event := range d.queue {
		for _, h := range d.handlers {
			done := make(chan struct{}, 1)
			go func(fn DispatchHandler, ev ChangeEvent) {
				_ = fn(ev)
				done <- struct{}{}
			}(h, event)
			select {
				case <-done:
				case <-time.After(d.policy.Timeout):
			}
		}
	}
}
