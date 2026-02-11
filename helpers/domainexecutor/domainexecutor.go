package domainexecutor

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Task func(ctx context.Context) error

var (
	ErrExecutorClosed = errors.New("domain executor closed")
	ErrDomainClosed   = errors.New("domain closed")
	ErrTimeout        = errors.New("timeout waiting for execution")
)

type Config struct {
	MaxWait        time.Duration
	IdleEvictAfter time.Duration
	QueueCapacity  int
}

type DomainExecutor struct {
	mu      sync.Mutex
	runners map[string]*domainRunner
	cfg     Config

	stopped bool
	stopCh  chan struct{}

	wgTasks   sync.WaitGroup
	wgDomains sync.WaitGroup
}

type domainRunner struct {
	queue chan request
	stop  chan struct{}
}

type request struct {
	ctx  context.Context
	task Task
	done chan error
}

func New(cfg Config) *DomainExecutor {
	if cfg.QueueCapacity <= 0 {
		cfg.QueueCapacity = 1
	}

	return &DomainExecutor{
		runners: make(map[string]*domainRunner),
		cfg:     cfg,
		stopCh:  make(chan struct{}),
	}
}

func NewDefault() *DomainExecutor {
	return &DomainExecutor{
		runners: make(map[string]*domainRunner),
		cfg: Config{
			QueueCapacity:  1,
			MaxWait:        10 * time.Second,
			IdleEvictAfter: time.Minute,
		},
		stopCh: make(chan struct{}),
	}
}

func (e *DomainExecutor) Execute(ctx context.Context, domain string, task Task) error {
	runner, err := e.getOrCreate(domain)
	if err != nil {
		return err
	}

	waitCtx := ctx
	if e.cfg.MaxWait > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, e.cfg.MaxWait)
		defer cancel()
	}

	req := request{
		ctx:  waitCtx,
		task: task,
		done: make(chan error, 1),
	}

	select {
	case <-e.stopCh:
		return ErrExecutorClosed
	default:
	}

	select {
	case <-runner.stop:
		return ErrDomainClosed
	default:
	}

	e.wgTasks.Add(1)

	select {
	case runner.queue <- req:
	case <-waitCtx.Done():
		e.wgTasks.Done()
		return ErrTimeout
	case <-e.stopCh:
		e.wgTasks.Done()
		return ErrExecutorClosed
	case <-runner.stop:
		e.wgTasks.Done()
		return ErrDomainClosed
	}

	select {
	case err := <-req.done:
		return err
	case <-waitCtx.Done():
		return ErrTimeout
	}
}

func (e *DomainExecutor) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	if e.stopped {
		e.mu.Unlock()
		return ErrExecutorClosed
	}

	e.stopped = true
	close(e.stopCh)

	runners := make([]*domainRunner, 0, len(e.runners))
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

func (e *DomainExecutor) getOrCreate(domain string) (*domainRunner, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		return nil, ErrExecutorClosed
	}

	if r, ok := e.runners[domain]; ok {
		return r, nil
	}

	r := &domainRunner{
		queue: make(chan request, e.cfg.QueueCapacity),
		stop:  make(chan struct{}),
	}

	e.runners[domain] = r

	e.wgDomains.Add(1)
	go e.runRunner(domain, r)

	return r, nil
}

func (e *DomainExecutor) runRunner(domain string, r *domainRunner) {
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

	resetTimer()

	for {
		select {
		case <-r.stop:
			e.cancelQueued(r)
			return

		case req := <-r.queue:
			resetTimer()
			err := req.task(req.ctx)
			req.done <- err
			e.wgTasks.Done()

		case <-timerCh:
			e.tryEvict(domain, r)
			resetTimer()
		}
	}
}

func (e *DomainExecutor) cancelQueued(r *domainRunner) {
	for {
		select {
		case req := <-r.queue:
			req.done <- ErrDomainClosed
			e.wgTasks.Done()
		default:
			return
		}
	}
}

func (e *DomainExecutor) tryEvict(domain string, r *domainRunner) {
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

func (e *DomainExecutor) stopRunner(r *domainRunner) {
	select {
	case <-r.stop:
	default:
		close(r.stop)
	}
}
