package domainexecutor

import (
	"context"
	"sync"
	"time"
)

type TryExecutor struct {
	mu      sync.Mutex
	runners map[string]*tryRunner
	cfg     TryConfig

	stopped bool
	stopCh  chan struct{}

	wgTasks   sync.WaitGroup
	wgDomains sync.WaitGroup
}

type TryConfig struct {
	IdleEvictAfter time.Duration
	QueueCapacity  int
}

type tryRunner struct {
	queue chan tryRequest
	stop  chan struct{}
}

type tryRequest struct {
	ctx    context.Context
	cancel context.CancelFunc
	task   Task
	cb     StateCallback
}

func NewTry(cfg TryConfig) *TryExecutor {
	if cfg.QueueCapacity <= 0 {
		cfg.QueueCapacity = 1
	}

	return &TryExecutor{
		runners: make(map[string]*tryRunner),
		cfg:     cfg,
		stopCh:  make(chan struct{}),
	}
}

func NewTryDefault() *TryExecutor {
	return NewTry(TryConfig{
		IdleEvictAfter: time.Minute,
		QueueCapacity:  1,
	})
}

func (e *TryExecutor) TryExecute(ctx context.Context, domain string, task Task, cb StateCallback) (bool, error) {
	if task == nil {
		return false, nil
	}

	runner, err := e.getOrCreate(domain)
	if err != nil {
		return false, err
	}

	reqCtx, cancel := context.WithCancel(ctx)
	req := tryRequest{
		ctx:    reqCtx,
		cancel: cancel,
		task:   task,
		cb:     cb,
	}

	e.wgTasks.Add(1)

	select {
	case runner.queue <- req:
		if req.cb != nil {
			req.cb(StatePending, nil)
		}
		return true, nil
	case <-e.stopCh:
		cancel()
		e.wgTasks.Done()
		return false, ErrExecutorClosed
	case <-runner.stop:
		cancel()
		e.wgTasks.Done()
		return false, ErrDomainClosed
	default:
		cancel()
		e.wgTasks.Done()
		return false, nil
	}
}

func (e *TryExecutor) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	if e.stopped {
		e.mu.Unlock()
		return ErrExecutorClosed
	}

	e.stopped = true
	close(e.stopCh)

	runners := make([]*tryRunner, 0, len(e.runners))
	for _, r := range e.runners {
		runners = append(runners, r)
	}
	e.mu.Unlock()

	for _, r := range runners {
		e.stopRunner(r)
	}

	done := make(chan struct{})
	go func() {
		e.wgTasks.Wait()
		e.wgDomains.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *TryExecutor) getOrCreate(domain string) (*tryRunner, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		return nil, ErrExecutorClosed
	}

	if r, ok := e.runners[domain]; ok {
		return r, nil
	}

	r := &tryRunner{
		queue: make(chan tryRequest, e.cfg.QueueCapacity),
		stop:  make(chan struct{}),
	}

	e.runners[domain] = r
	e.wgDomains.Add(1)
	go e.runRunner(domain, r)

	return r, nil
}

func (e *TryExecutor) runRunner(domain string, r *tryRunner) {
	defer e.wgDomains.Done()

	var timer *time.Timer
	var timerCh <-chan time.Time

	resetTimer := func() {
		if e.cfg.IdleEvictAfter <= 0 {
			return
		}

		if timer == nil {
			timer = time.NewTimer(e.cfg.IdleEvictAfter)
			timerCh = timer.C
			return
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		timer.Reset(e.cfg.IdleEvictAfter)
		timerCh = timer.C
	}

	stopTimer := func() {
		if timer == nil {
			return
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}

	resetTimer()

	for {
		select {
		case <-r.stop:
			stopTimer()
			e.drainQueued(r)
			return

		case req := <-r.queue:
			select {
			case <-r.stop:
				e.cancelRequest(req, ErrDomainClosed)
				stopTimer()
				e.drainQueued(r)
				return
			default:
			}

			resetTimer()

			if req.cb != nil && req.ctx.Err() == nil {
				req.cb(StateRunning, nil)
			}

			err := req.task(req.ctx)

			if req.cb != nil {
				switch {
				case req.ctx.Err() != nil:
					req.cb(StateCancelled, req.ctx.Err())
				case err != nil:
					req.cb(StateFailed, err)
				default:
					req.cb(StateCompleted, nil)
				}
			}

			e.wgTasks.Done()

			select {
			case <-r.stop:
				stopTimer()
				e.drainQueued(r)
				return
			default:
			}

		case <-timerCh:
			e.tryEvict(domain, r)
			resetTimer()
		}
	}
}

func (e *TryExecutor) drainQueued(r *tryRunner) {
	for {
		select {
		case req := <-r.queue:
			e.cancelRequest(req, ErrDomainClosed)
		default:
			return
		}
	}
}

func (e *TryExecutor) cancelRequest(req tryRequest, err error) {
	req.cancel()
	if req.cb != nil {
		req.cb(StateCancelled, err)
	}
	e.wgTasks.Done()
}

func (e *TryExecutor) tryEvict(domain string, r *tryRunner) {
	select {
	case <-e.stopCh:
		return
	default:
	}

	e.mu.Lock()
	current, ok := e.runners[domain]
	if ok && current == r {
		delete(e.runners, domain)
	}
	e.mu.Unlock()

	if ok && current == r {
		e.stopRunner(r)
	}
}

func (e *TryExecutor) stopRunner(r *tryRunner) {
	select {
	case <-r.stop:
	default:
		close(r.stop)
	}
}
