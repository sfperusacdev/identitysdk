package domainexecutor

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSerialExecutionPerDomain(t *testing.T) {
	exec := New(Config{
		MaxWait:       2 * time.Second,
		QueueCapacity: 1,
	})

	var running int32
	errCh := make(chan error, 1)

	task := func(ctx context.Context) error {
		if atomic.AddInt32(&running, 1) != 1 {
			select {
			case errCh <- errors.New("more than one running concurrently"):
			default:
			}
		}

		time.Sleep(30 * time.Millisecond)
		atomic.AddInt32(&running, -1)
		return nil
	}

	wg := sync.WaitGroup{}
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = exec.Execute(context.Background(), "a", task)
		}()
	}

	wg.Wait()

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestParallelAcrossDomains(t *testing.T) {
	exec := New(Config{
		MaxWait:       2 * time.Second,
		QueueCapacity: 1,
	})

	start := time.Now()

	task := func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = exec.Execute(context.Background(), "a", task)
	}()

	go func() {
		defer wg.Done()
		_ = exec.Execute(context.Background(), "b", task)
	}()

	wg.Wait()

	if time.Since(start) > 350*time.Millisecond {
		t.Fatal("domains did not run in parallel")
	}
}

func TestTimeoutWaitingTurn(t *testing.T) {
	exec := New(Config{
		MaxWait:       100 * time.Millisecond,
		QueueCapacity: 1,
	})

	block := make(chan struct{})

	go func() {
		_ = exec.Execute(context.Background(), "a", func(ctx context.Context) error {
			<-block
			return nil
		})
	}()

	time.Sleep(20 * time.Millisecond)

	err := exec.Execute(context.Background(), "a", func(ctx context.Context) error {
		return nil
	})

	if err != ErrTimeout {
		t.Fatalf("expected timeout got %v", err)
	}

	close(block)
}

func TestIdleEviction(t *testing.T) {
	exec := New(Config{
		MaxWait:        time.Second,
		IdleEvictAfter: 50 * time.Millisecond,
		QueueCapacity:  1,
	})

	if err := exec.Execute(context.Background(), "x", func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(120 * time.Millisecond)

	if err := exec.Execute(context.Background(), "x", func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestShutdownGracefulWaits(t *testing.T) {
	exec := New(Config{
		MaxWait:       time.Second,
		QueueCapacity: 1,
	})

	started := make(chan struct{})
	finish := make(chan struct{})

	go func() {
		_ = exec.Execute(context.Background(), "a", func(ctx context.Context) error {
			close(started)
			<-finish
			return nil
		})
	}()

	<-started

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

func TestHighConcurrencySameDomain(t *testing.T) {
	exec := New(Config{
		MaxWait:       5 * time.Second,
		QueueCapacity: 1,
	})

	var running int32
	errCh := make(chan error, 1)

	task := func(ctx context.Context) error {
		if atomic.AddInt32(&running, 1) != 1 {
			select {
			case errCh <- errors.New("more than one running concurrently"):
			default:
			}
		}

		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&running, -1)
		return nil
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = exec.Execute(context.Background(), "z", task)
		}()
	}

	wg.Wait()

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestTimeoutWhileWaitingTurn(t *testing.T) {
	exec := New(Config{
		MaxWait:       50 * time.Millisecond,
		QueueCapacity: 1,
	})

	block := make(chan struct{})

	go func() {
		_ = exec.Execute(context.Background(), "a", func(ctx context.Context) error {
			<-block
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	err := exec.Execute(context.Background(), "a", func(ctx context.Context) error {
		return nil
	})

	if err != ErrTimeout {
		t.Fatalf("expected ErrTimeout got %v", err)
	}

	close(block)
}

func TestTimeoutDuringExecution(t *testing.T) {
	exec := New(Config{
		MaxWait:       30 * time.Millisecond,
		QueueCapacity: 1,
	})

	err := exec.Execute(context.Background(), "a", func(ctx context.Context) error {
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestContextCancelStopsWaiting(t *testing.T) {
	exec := New(Config{
		MaxWait:       time.Second,
		QueueCapacity: 1,
	})

	block := make(chan struct{})

	go func() {
		_ = exec.Execute(context.Background(), "a", func(ctx context.Context) error {
			<-block
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := exec.Execute(ctx, "a", func(ctx context.Context) error {
		return nil
	})

	if err == nil {
		t.Fatal("expected cancel error")
	}

	close(block)
}

func TestShutdownRejectsNewExecution(t *testing.T) {
	exec := New(Config{
		MaxWait:       time.Second,
		QueueCapacity: 1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = exec.Shutdown(context.Background())

	err := exec.Execute(ctx, "x", func(ctx context.Context) error {
		return nil
	})

	if err != ErrExecutorClosed {
		t.Fatalf("expected ErrExecutorClosed got %v", err)
	}
}

func TestShutdownWaitsForInflight(t *testing.T) {
	exec := New(Config{
		MaxWait:       time.Second,
		QueueCapacity: 1,
	})

	start := make(chan struct{})
	finish := make(chan struct{})

	go func() {
		_ = exec.Execute(context.Background(), "a", func(ctx context.Context) error {
			close(start)
			<-finish
			return nil
		})
	}()

	<-start

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		_ = exec.Shutdown(ctx)
		close(done)
	}()

	time.Sleep(40 * time.Millisecond)

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

func TestDifferentDomainsDoNotBlock(t *testing.T) {
	exec := New(Config{
		MaxWait:       time.Second,
		QueueCapacity: 1,
	})

	start := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(2)

	task := func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	}

	go func() {
		defer wg.Done()
		_ = exec.Execute(context.Background(), "a", task)
	}()

	go func() {
		defer wg.Done()
		_ = exec.Execute(context.Background(), "b", task)
	}()

	wg.Wait()

	if time.Since(start) > 280*time.Millisecond {
		t.Fatal("domains executed sequentially, expected parallel")
	}
}

func TestWaitsStrictlyOneAtTime(t *testing.T) {
	exec := New(Config{
		MaxWait:       5 * time.Second,
		QueueCapacity: 1,
	})

	var running int32
	errCh := make(chan error, 1)

	task := func(ctx context.Context) error {
		if atomic.AddInt32(&running, 1) != 1 {
			select {
			case errCh <- errors.New("parallel execution detected"):
			default:
			}
		}

		time.Sleep(15 * time.Millisecond)
		atomic.AddInt32(&running, -1)
		return nil
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = exec.Execute(context.Background(), "k", task)
		}()
	}

	wg.Wait()

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}
