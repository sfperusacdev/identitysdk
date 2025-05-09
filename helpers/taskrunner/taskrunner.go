package taskrunner

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type TaskRunner struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning atomic.Bool
	task      func(context.Context)
	interval  time.Duration
}

func NewTaskRunner(task func(context.Context), interval time.Duration) *TaskRunner {
	return &TaskRunner{
		task:     task,
		interval: interval,
	}
}

func (tr *TaskRunner) Start() {
	if tr.isRunning.Load() {
		return
	}

	tr.isRunning.Store(true)
	tr.ctx, tr.cancel = context.WithCancel(context.Background())
	tr.wg.Add(1)

	go func() {
		defer func() {
			tr.isRunning.Store(false)
			tr.wg.Done()
		}()
		for {
			select {
			case <-tr.ctx.Done():
				return
			default:
				tr.task(tr.ctx)
			}

			select {
			case <-tr.ctx.Done():
				return
			case <-time.After(tr.interval):
			}
		}
	}()
}

func (tr *TaskRunner) Stop() {
	if !tr.isRunning.Load() {
		return
	}

	tr.cancel()
	tr.wg.Wait()
}

func (tr *TaskRunner) IsRunning() bool {
	return tr.isRunning.Load()
}
