package events

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueName string

type QueueConnection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func connectRabbitMQ(url string) (*QueueConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	slog.Info("connected to RabbitMQ")
	return &QueueConnection{Conn: conn, Channel: ch}, nil
}

func closeQueueConnection(qc *QueueConnection) error {
	if qc == nil {
		return nil
	}
	var errs []error
	if qc.Channel != nil {
		if err := qc.Channel.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			errs = append(errs, err)
		}
	}
	if qc.Conn != nil {
		if err := qc.Conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

const (
	backoffStart = 10 * time.Second
	backoffMax   = 1 * time.Minute
)

func withJitter(d time.Duration) time.Duration {
	bound := int64(d / 5)
	if bound < 1 {
		return d
	}
	n := time.Duration(rand.Int63n(bound))
	if rand.Intn(2) == 0 {
		return d - n
	}
	return d + n
}

func reconnect(ctx context.Context, url string) (*QueueConnection, error) {
	delay := backoffStart
	for {
		qc, err := connectRabbitMQ(url)
		if err == nil {
			return qc, nil
		}
		slog.Warn("reconnection attempt failed", "error", err, "retry_in", delay)
		timer := time.NewTimer(withJitter(delay))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, ctx.Err()
		case <-timer.C:
		}
		if delay < backoffMax/2 {
			delay *= 2
		} else {
			delay = backoffMax
		}
	}
}

func errOrClosed(err *amqp.Error) error {
	if err == nil {
		return amqp.ErrClosed
	}
	return err
}
