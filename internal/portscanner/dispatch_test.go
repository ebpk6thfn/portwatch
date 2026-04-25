package portscanner

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func makeDispatchEvent(port uint16, kind string) ChangeEvent {
	return ChangeEvent{
		Kind: kind,
		Entry: Entry{
			Port:     port,
			Protocol: "tcp",
			Addr:     "0.0.0.0",
		},
	}
}

func TestDispatcher_DefaultPolicy(t *testing.T) {
	p := DefaultDispatchPolicy()
	if p.Workers <= 0 {
		t.Fatalf("expected positive workers, got %d", p.Workers)
	}
	if p.QueueDepth <= 0 {
		t.Fatalf("expected positive queue depth, got %d", p.QueueDepth)
	}
	if p.Timeout <= 0 {
		t.Fatalf("expected positive timeout, got %v", p.Timeout)
	}
}

func TestDispatcher_HandlerCalledForEvent(t *testing.T) {
	d := NewDispatcher(DispatchPolicy{Workers: 2, QueueDepth: 16, Timeout: time.Second})
	defer d.Close()

	var called int32
	var wg sync.WaitGroup
	wg.Add(1)
	d.Register(func(ev ChangeEvent) error {
		atomic.AddInt32(&called, 1)
		wg.Done()
		return nil
	})

	ok := d.Dispatch(makeDispatchEvent(8080, EventOpened))
	if !ok {
		t.Fatal("expected Dispatch to return true")
	}
	wg.Wait()
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected handler called once, got %d", called)
	}
}

func TestDispatcher_MultipleHandlersCalled(t *testing.T) {
	d := NewDispatcher(DispatchPolicy{Workers: 2, QueueDepth: 16, Timeout: time.Second})
	defer d.Close()

	var count int32
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		d.Register(func(ev ChangeEvent) error {
			atomic.AddInt32(&count, 1)
			wg.Done()
			return nil
		})
	}

	d.Dispatch(makeDispatchEvent(9090, EventOpened))
	wg.Wait()
	if atomic.LoadInt32(&count) != 2 {
		t.Fatalf("expected 2 handler calls, got %d", count)
	}
}

func TestDispatcher_FullQueue_ReturnsFalse(t *testing.T) {
	// Use a tiny queue with no workers reading from it.
	d := &Dispatcher{
		policy:   DispatchPolicy{Workers: 0, QueueDepth: 1, Timeout: time.Second},
		queue:    make(chan ChangeEvent, 1),
		stop:     make(chan struct{}),
	}

	ok1 := d.Dispatch(makeDispatchEvent(1111, EventOpened))
	if !ok1 {
		t.Fatal("expected first dispatch to succeed")
	}
	ok2 := d.Dispatch(makeDispatchEvent(2222, EventOpened))
	if ok2 {
		t.Fatal("expected second dispatch to fail on full queue")
	}
}

func TestDispatcher_CloseIsIdempotent(t *testing.T) {
	d := NewDispatcher(DefaultDispatchPolicy())
	d.Close()
	d.Close() // should not panic
}
