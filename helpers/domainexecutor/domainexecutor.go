// Package domainexecutor proporciona un ejecutor de tareas concurrente que garantiza
// ejecución serial por dominio.
//
// Cada dominio tiene un "runner" dedicado que procesa sus tareas una a una en orden.
// Esto asegura que dos tareas del mismo dominio nunca se ejecuten en paralelo,
// mientras que tareas de dominios distintos sí pueden ejecutarse concurrentemente.
//
// Funcionamiento general:
//
// 1. Cuando se llama Execute(domain, task):
//   - Se obtiene o crea un runner para ese dominio.
//   - La tarea se encola en la cola del dominio.
//   - El runner ejecuta las tareas secuencialmente.
//
// 2. Concurrencia:
//   - Diferentes dominios se ejecutan en paralelo.
//   - Dentro de un mismo dominio las tareas se ejecutan estrictamente una a la vez.
//
// 3. Control de cola:
//   - QueueCapacity limita cuántas tareas pueden esperar por dominio.
//   - Si la cola está llena, Execute espera hasta MaxWait para poder encolar.
//   - Si el tiempo se supera, devuelve ErrTimeout.
//
// 4. Estados de tarea:
//   - Se puede registrar un callback para recibir cambios de estado:
//     pending → running → completed/failed/timeout/cancelled.
//
// 5. Evicción de dominios inactivos:
//   - Si un dominio no recibe tareas durante IdleEvictAfter,
//     su runner se elimina automáticamente para liberar recursos.
//
// 6. Shutdown:
//   - Shutdown detiene el executor.
//   - Cancela tareas en cola.
//   - Espera a que las tareas en ejecución finalicen.
//
// Este patrón permite implementar procesamiento seguro por clave (dominio),
// evitando condiciones de carrera cuando múltiples operaciones afectan
// el mismo recurso lógico.
package domainexecutor

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Task func(ctx context.Context) error
type TaskState string

const (
	StatePending   TaskState = "pending"
	StateRunning   TaskState = "running"
	StateCompleted TaskState = "completed"
	StateFailed    TaskState = "failed"
	StateTimeout   TaskState = "timeout"
	StateCancelled TaskState = "cancelled"
)

type StateCallback func(state TaskState, err error)

var (
	ErrExecutorClosed = errors.New("domain executor closed")
	ErrDomainClosed   = errors.New("domain closed")
	ErrTimeout        = errors.New("timeout waiting for execution")
)

type Config struct {
	// MaxWait tiempo máximo que Execute espera para:
	// 1) poder poner la tarea en la cola del dominio
	// 2) recibir el resultado de la ejecución
	// si ese tiempo se supera, Execute devuelve ErrTimeout
	MaxWait time.Duration

	// IdleEvictAfter tiempo que un runner de dominio puede estar
	// inactivo antes de ser eliminado
	IdleEvictAfter time.Duration

	// QueueCapacity número máximo de tareas que pueden quedar
	// esperando en la cola por dominio
	QueueCapacity int
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
	cb   StateCallback
}

func NewDefault() *DomainExecutor {
	return New(Config{
		MaxWait:        10 * time.Second,
		IdleEvictAfter: time.Minute,
		QueueCapacity:  1,
	})
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

// Execute encola y ejecuta una tarea asociada a un dominio.
// Devuelve:
// - ErrExecutorClosed si el executor ya fue cerrado
// - ErrDomainClosed si el runner del dominio fue cerrado
// - ErrTimeout si se supera MaxWait esperando encolar o esperando el resultado
// - error retornado por la Task si la ejecución falla
// - nil si la Task termina correctamente
func (e *DomainExecutor) Execute(ctx context.Context, domain string, task Task, cb StateCallback) error {
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
		cb:   cb,
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

		if req.cb != nil {
			req.cb(StatePending, nil)
		}

	case <-waitCtx.Done():
		e.wgTasks.Done()
		if req.cb != nil {
			req.cb(StateTimeout, waitCtx.Err())
		}
		return ErrTimeout

	case <-e.stopCh:
		e.wgTasks.Done()
		if req.cb != nil {
			req.cb(StateCancelled, ErrExecutorClosed)
		}
		return ErrExecutorClosed

	case <-runner.stop:
		e.wgTasks.Done()
		if req.cb != nil {
			req.cb(StateCancelled, ErrDomainClosed)
		}
		return ErrDomainClosed
	}

	select {
	case err := <-req.done:
		return err

	case <-waitCtx.Done():
		if req.cb != nil {
			req.cb(StateTimeout, waitCtx.Err())
		}
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

			if req.cb != nil && req.ctx.Err() == nil {
				req.cb(StateRunning, nil)
			}

			err := req.task(req.ctx)

			if req.ctx.Err() == nil {
				if req.cb != nil {
					if err != nil {
						req.cb(StateFailed, err)
					} else {
						req.cb(StateCompleted, nil)
					}
				}
			}

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
			if req.cb != nil {
				req.cb(StateCancelled, ErrDomainClosed)
			}
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
