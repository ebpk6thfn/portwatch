package portscanner

import (
	"context"
	"time"
)

// Flusher periodically drains a RingBuffer and dispatches batched events
// to a handler function. Useful for coalescing bursts before notification.
type Flusher struct {
	buf      *RingBuffer
	interval time.Duration
	handler  func([]ChangeEvent)
}

// NewFlusher creates a Flusher that drains buf every interval and calls handler.
func NewFlusher(buf *RingBuffer, interval time.Duration, handler func([]ChangeEvent)) *Flusher {
	return &Flusher{
		buf:      buf,
		interval: interval,
		handler:  handler,
	}
}

// Run starts the flush loop. It blocks until ctx is cancelled.
func (f *Flusher) Run(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// final flush on shutdown
			if events := f.buf.Drain(); len(events) > 0 {
				f.handler(events)
			}
			return
		case <-ticker.C:
			if events := f.buf.Drain(); len(events) > 0 {
				f.handler(events)
			}
		}
	}
}
