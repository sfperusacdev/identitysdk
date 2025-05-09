package taskrunner_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/helpers/taskrunner"
)

func TestConcurrentStartStop(t *testing.T) {

	task := func(cxt context.Context) {
		fmt.Println("Executing task at", time.Now().Format(time.RFC3339))
	}

	runner := taskrunner.NewTaskRunner(task, 2*time.Second)

	var wg sync.WaitGroup

	for range 10 {
		wg.Add(2)

		go func() {
			defer wg.Done()
			runner.Start()
		}()

		go func() {
			defer wg.Done()
			time.Sleep(500 * time.Millisecond)
			runner.Stop()
		}()
	}

	wg.Wait()
}

func TestTaskExecution(t *testing.T) {
	var taskCount int64
	expectedExecutions := 5
	interval := 500 * time.Millisecond
	totalWaitTime := time.Duration(expectedExecutions) * interval

	task := func(cxt context.Context) {
		atomic.AddInt64(&taskCount, 1)
	}

	runner := taskrunner.NewTaskRunner(task, interval)

	startTime := time.Now()
	runner.Start()
	time.Sleep(totalWaitTime)
	runner.Stop()
	elapsedTime := time.Since(startTime)

	executions := atomic.LoadInt64(&taskCount)

	if executions < int64(expectedExecutions) {
		t.Fatalf("Expected at least %d executions, but got %d", expectedExecutions, executions)
	}

	if elapsedTime < totalWaitTime {
		t.Fatalf("Expected at least %v elapsed, but got %v", totalWaitTime, elapsedTime)
	}

	t.Logf("Task executed %d times in %v", executions, elapsedTime)
}
