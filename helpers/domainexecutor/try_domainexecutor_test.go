package domainexecutor

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTryExecutorSerialExecutionPerDomain(t *testing.T) {
	const total = 20

	exec := NewTry(TryConfig{
		QueueCapacity: total,
	})

	var running int32
	errCh := make(chan error, 1)
	done := make(chan struct{}, total)

	task := func(ctx context.Context) error {
		if atomic.AddInt32(&running, 1) != 1 {
			select {
			case errCh <- errors.New("parallel execution detected"):
			default:
			}
		}

		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&running, -1)
		done <- struct{}{}
		return nil
	}

	for i := 0; i < total; i++ {
		ok, err := exec.TryExecute(context.Background(), "a", task, nil)
		if !ok || err != nil {
			t.Fatalf("enqueue failed at %d: ok=%v err=%v", i, ok, err)
		}
	}

	for i := 0; i < total; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("tasks did not finish")
		}
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestTryExecutorParallelAcrossDomains(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	var wg sync.WaitGroup
	wg.Add(2)

	task := func(ctx context.Context) error {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	start := time.Now()

	ok, err := exec.TryExecute(context.Background(), "a", task, nil)
	if !ok || err != nil {
		t.Fatalf("enqueue a failed: ok=%v err=%v", ok, err)
	}

	ok, err = exec.TryExecute(context.Background(), "b", task, nil)
	if !ok || err != nil {
		t.Fatalf("enqueue b failed: ok=%v err=%v", ok, err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("tasks did not finish")
	}

	if time.Since(start) > 350*time.Millisecond {
		t.Fatal("domains did not run in parallel")
	}
}

func TestTryExecutorQueueFullReturnsFalse(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	block := make(chan struct{})
	started := make(chan struct{})

	ok, err := exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		close(started)
		<-block
		return nil
	}, nil)
	if !ok || err != nil {
		t.Fatal("first enqueue should succeed")
	}

	<-started

	ok, err = exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		<-block
		return nil
	}, nil)
	if !ok || err != nil {
		t.Fatal("second enqueue should succeed")
	}

	ok, err = exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		return nil
	}, nil)
	if ok {
		t.Fatal("expected enqueue to fail when queue is full")
	}
	if err != nil {
		t.Fatalf("expected nil err got %v", err)
	}

	close(block)
}

func TestTryExecutorIdleEviction(t *testing.T) {
	exec := NewTry(TryConfig{
		IdleEvictAfter: 50 * time.Millisecond,
		QueueCapacity:  1,
	})

	done := make(chan struct{})

	ok, err := exec.TryExecute(context.Background(), "x", func(ctx context.Context) error {
		close(done)
		return nil
	}, nil)
	if !ok || err != nil {
		t.Fatal("first execution failed")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("first task did not finish")
	}

	time.Sleep(120 * time.Millisecond)

	exec.mu.Lock()
	_, exists := exec.runners["x"]
	exec.mu.Unlock()

	if exists {
		t.Fatal("runner should have been evicted")
	}

	done2 := make(chan struct{})

	ok, err = exec.TryExecute(context.Background(), "x", func(ctx context.Context) error {
		close(done2)
		return nil
	}, nil)
	if !ok || err != nil {
		t.Fatal("execution after eviction failed")
	}

	select {
	case <-done2:
	case <-time.After(time.Second):
		t.Fatal("second task did not finish")
	}
}

func TestTryExecutorShutdownWaitsInflight(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	start := make(chan struct{})
	finish := make(chan struct{})

	ok, err := exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		close(start)
		<-finish
		return nil
	}, nil)
	if !ok || err != nil {
		t.Fatal("task should be accepted")
	}

	<-start

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		_ = exec.Shutdown(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("shutdown returned before task finished")
	default:
	}

	close(finish)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not complete")
	}
}

func TestTryExecutorShutdownRejectsNewTasks(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	_ = exec.Shutdown(context.Background())

	ok, err := exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		return nil
	}, nil)

	if ok {
		t.Fatal("should not accept tasks after shutdown")
	}

	if err != ErrExecutorClosed {
		t.Fatalf("expected ErrExecutorClosed got %v", err)
	}
}

func TestTryExecutorStateCallback(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	var states []TaskState
	var mu sync.Mutex
	done := make(chan struct{})

	cb := func(s TaskState, err error) {
		mu.Lock()
		states = append(states, s)
		mu.Unlock()

		if s == StateCompleted {
			close(done)
		}
	}

	ok, err := exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		return nil
	}, cb)
	if !ok || err != nil {
		t.Fatal("task should be accepted")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected completed callback")
	}

	expected := []TaskState{
		StatePending,
		StateRunning,
		StateCompleted,
	}

	mu.Lock()
	defer mu.Unlock()

	if len(states) != len(expected) {
		t.Fatalf("expected %v got %v", expected, states)
	}

	for i := range expected {
		if states[i] != expected[i] {
			t.Fatalf("expected %v got %v", expected, states)
		}
	}
}

func TestTryExecutorStateFailed(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	done := make(chan struct{})
	var last TaskState
	var mu sync.Mutex

	cb := func(s TaskState, err error) {
		mu.Lock()
		last = s
		mu.Unlock()

		if s == StateFailed {
			close(done)
		}
	}

	ok, err := exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		return errors.New("fail")
	}, cb)
	if !ok || err != nil {
		t.Fatal("task should be accepted")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected failed callback")
	}

	mu.Lock()
	defer mu.Unlock()

	if last != StateFailed {
		t.Fatalf("expected StateFailed got %v", last)
	}
}

func TestTryExecutorStateCancelledOnShutdown(t *testing.T) {
	exec := NewTry(TryConfig{
		QueueCapacity: 1,
	})

	block := make(chan struct{})
	started := make(chan struct{})
	cancelled := make(chan struct{})

	ok, err := exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		close(started)
		<-block
		return nil
	}, nil)
	if !ok || err != nil {
		t.Fatal("first task should run")
	}

	<-started

	cb := func(s TaskState, err error) {
		if s == StateCancelled {
			close(cancelled)
		}
	}

	ok, err = exec.TryExecute(context.Background(), "a", func(ctx context.Context) error {
		return nil
	}, cb)
	if !ok || err != nil {
		t.Fatal("second task should enqueue")
	}

	done := make(chan struct{})
	go func() {
		_ = exec.Shutdown(context.Background())
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	close(block)

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("expected StateCancelled")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not complete")
	}
}
