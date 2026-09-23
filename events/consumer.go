package events

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sfperusacdev/identitysdk/configs"
	"go.uber.org/fx"
)

type Consumer interface {
	Queue() QueueName
	Handle(context.Context, []byte) error
}

const ConsumerGroupTag = `group:"events.consumers"`

func AsConsumer(fn any) any {
	return fx.Annotate(fn, fx.ResultTags(ConsumerGroupTag))
}

type consumerList []Consumer

func mapConsumers(consumers []Consumer) consumerList {
	if len(consumers) == 0 {
		slog.Warn("no event consumers registered")
	} else {
		slog.Info("event consumers registered", "count", len(consumers))
	}
	return consumerList(consumers)
}

func StartEventBusConsumers(lc fx.Lifecycle, config configs.GeneralServiceConfigProvider, consumers consumerList) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			for _, consumer := range consumers {
				wg.Go(func() { runConsumer(ctx, config.RabbitMQURL(), consumer) })
			}
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			wg.Wait()
			return nil
		},
	})
}

func runConsumer(ctx context.Context, url string, consumer Consumer) {
	queue := consumer.Queue()
	if queue == "" {
		slog.Error("event consumer has empty queue")
		return
	}
	for {
		qc, err := reconnect(ctx, url)
		if err != nil {
			return
		}
		err = consumeConnection(ctx, qc, consumer)
		_ = closeQueueConnection(qc)
		if ctx.Err() != nil {
			return
		}
		slog.Warn("event consumer restarting", "queue", queue, "error", err)
	}
}

func consumeConnection(ctx context.Context, qc *QueueConnection, consumer Consumer) error {
	queue := consumer.Queue()
	if _, err := qc.Channel.QueueDeclare(string(queue), true, false, false, false, nil); err != nil {
		return err
	}
	msgs, err := qc.Channel.Consume(string(queue), "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	notifyClose := qc.Conn.NotifyClose(make(chan *amqp.Error, 1))
	notifyChannelClose := qc.Channel.NotifyClose(make(chan *amqp.Error, 1))
	slog.Info("started consuming events", "queue", queue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-notifyClose:
			return errOrClosed(err)
		case err := <-notifyChannelClose:
			return errOrClosed(err)
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("consumer delivery channel closed")
			}
			if err := consumer.Handle(ctx, msg.Body); err != nil {
				slog.Error("event handler failed", "queue", queue, "error", err)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					return nackErr
				}
				continue
			}
			if err := msg.Ack(false); err != nil {
				return err
			}
		}
	}
}
