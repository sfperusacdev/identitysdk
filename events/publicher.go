package events

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sfperusacdev/identitysdk/configs"
	"go.uber.org/fx"
)

type Publisher interface {
	Publish(ctx context.Context, queue QueueName, body any) error
	Close() error
}

// PublicherChannel is kept for compatibility with the original misspelled API.
type PublicherChannel interface {
	Publisher
	Publich(ctx context.Context, queue QueueName, body any) error
}

type publisher struct {
	url string

	mu      sync.RWMutex
	opMu    sync.Mutex
	active  *QueueConnection
	queues  map[QueueName]bool
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	started bool
	closed  bool
}

func NewPublisher(lc fx.Lifecycle, config configs.GeneralServiceConfigProvider) Publisher {
	p := newPublisher(config.RabbitMQURL())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			p.start()
			return nil
		},
		OnStop: func(context.Context) error { return p.Close() },
	})
	return p
}

func NewPublicherChannel(config configs.GeneralServiceConfigProvider) PublicherChannel {
	p := newPublisher(config.RabbitMQURL())
	p.start()
	return p
}

func newPublisher(url string) *publisher {
	ctx, cancel := context.WithCancel(context.Background())
	return &publisher{
		url:    url,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
		queues: make(map[QueueName]bool),
	}
}

func (p *publisher) start() {
	p.mu.Lock()
	if p.started || p.closed {
		p.mu.Unlock()
		return
	}
	p.started = true
	p.mu.Unlock()
	go p.handleReconnection()
}

func (p *publisher) handleReconnection() {
	defer close(p.done)
	defer p.clearConnection(nil)

	for {
		qc, err := reconnect(p.ctx, p.url)
		if err != nil {
			return
		}
		p.replaceConnection(qc)

		notifyClose := qc.Conn.NotifyClose(make(chan *amqp.Error, 1))
		notifyChannelClose := qc.Channel.NotifyClose(make(chan *amqp.Error, 1))
		select {
		case <-p.ctx.Done():
			return
		case err := <-notifyClose:
			slog.Warn("publisher connection closed", "error", err)
		case err := <-notifyChannelClose:
			slog.Warn("publisher channel closed", "error", err)
		}
		p.clearConnection(qc)
	}
}

func (p *publisher) replaceConnection(qc *QueueConnection) {
	p.opMu.Lock()
	p.mu.Lock()
	old := p.active
	p.active = qc
	p.queues = make(map[QueueName]bool)
	p.mu.Unlock()
	p.opMu.Unlock()
	closeQueueConnection(old)
	slog.Info("publisher connection updated")
}

func (p *publisher) clearConnection(expected *QueueConnection) {
	p.opMu.Lock()
	p.mu.Lock()
	if expected == nil || p.active == expected {
		old := p.active
		p.active = nil
		p.queues = make(map[QueueName]bool)
		p.mu.Unlock()
		p.opMu.Unlock()
		closeQueueConnection(old)
		return
	}
	p.mu.Unlock()
	p.opMu.Unlock()
}

func (p *publisher) connection() *QueueConnection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

func (p *publisher) declare(qc *QueueConnection, queue QueueName) error {
	if queue == "" {
		return errors.New("queue name is empty")
	}
	p.mu.RLock()
	declared := p.queues[queue]
	p.mu.RUnlock()
	if declared {
		return nil
	}
	if qc == nil || qc.Channel == nil {
		return amqp.ErrClosed
	}
	if _, err := qc.Channel.QueueDeclare(string(queue), true, false, false, false, nil); err != nil {
		return err
	}
	p.mu.Lock()
	p.queues[queue] = true
	p.mu.Unlock()
	slog.Info("queue declared", "queue", queue)
	return nil
}

func (p *publisher) Publish(ctx context.Context, queue QueueName, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	p.opMu.Lock()
	defer p.opMu.Unlock()
	qc := p.connection()
	if qc == nil || qc.Channel == nil {
		return amqp.ErrClosed
	}
	if err := p.declare(qc, queue); err != nil {
		return err
	}
	return qc.Channel.PublishWithContext(ctx, "", string(queue), false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (p *publisher) Publich(ctx context.Context, queue QueueName, data any) error {
	return p.Publish(ctx, queue, data)
}

func (p *publisher) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.cancel()
	started := p.started
	p.mu.Unlock()

	if started {
		<-p.done
	}
	p.opMu.Lock()
	p.mu.Lock()
	qc := p.active
	p.active = nil
	p.mu.Unlock()
	p.opMu.Unlock()
	return closeQueueConnection(qc)
}
