package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"email-sender/config"
	"email-sender/internal/handlers/rabbitmq"
	ctxlog "email-sender/internal/system/logger"
	"email-sender/internal/system/metrics"

	"github.com/streadway/amqp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Consumer interface {
	Consume()
	Close() error
}

type consumer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *zap.Logger
	metrics *metrics.Client
	conn    *amqp.Connection
	ch      *amqp.Channel
	connURL string
	handler rabbitmq.Handler
	queues  []*config.Queue
}

func NewConsumer(cfg *config.Consumer, handler rabbitmq.Handler, logger *zap.Logger, metrics *metrics.Client) (Consumer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = ctxlog.Enrich(ctx, logger)

	return &consumer{
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger,
		metrics: metrics,
		connURL: cfg.ConnectionURL,
		handler: handler,
		queues: []*config.Queue{
			//здесь добавление очередей из конфига
			cfg.NotificationsQueue,
		},
	}, nil
}

func (c *consumer) Consume() {
	go func() {
		for c.ctx.Err() == nil {
			if err := c.setupConnection(); err != nil {
				c.logger.With(zap.Error(err)).Error("failed to connect to RabbitMQ")
				// waiting one second to reconnect
				time.Sleep(time.Second)
				continue
			}
			c.logger.Info("successfully connected to RabbitMQ")

			if len(c.queues) == 0 {
				return
			}

			wg := &sync.WaitGroup{}
			for _, q := range c.queues {
				if q.Name == "" {
					continue
				}

				err := c.ch.ExchangeDeclare(q.Exchange, amqp.ExchangeFanout, q.ExchangeDurable, false, false, false, nil)
				if err != nil {
					c.logger.With(zap.Error(err)).Error("failed to declare a exchange")
					continue
				}

				queue, err := c.ch.QueueDeclare(q.Name, q.Durable, false, false, false, nil)
				if err != nil {
					c.logger.With(zap.Error(err)).Error("failed to declare a queue")
					continue
				}

				if err := c.ch.QueueBind(q.Name, q.RoutingKey, q.Exchange, false, nil); err != nil {
					c.logger.With(zap.Error(err)).Error("failed to bind a queue")
					continue
				}

				if err := c.ch.Qos(1, 0, false); err != nil {
					c.logger.With(zap.Error(err)).Error("failed to configure Qos")
					continue
				}

				messages, err := c.ch.Consume(queue.Name, "", false, false, false, false, nil)
				if err != nil {
					c.logger.With(zap.Error(err)).Error("failed to register a consumer")
					continue
				}

				wg.Add(1)
				go func() {
					for message := range messages {
						c.handler.Handle(c.ctx, queue.Name, message)
						if err := message.Ack(false); err != nil {
							c.logger.With(zap.Error(err)).Error("failed to acknowledge a message")
						}
					}
					wg.Done()
				}()
			}

			wg.Wait()
			c.logger.Info("rabbitMQ connection was closed")
		}
	}()
}

func (c *consumer) setupConnection() error {
	conn, err := amqp.Dial(c.connURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a RabbitMQ channel: %w", err)
	}

	c.conn = conn
	c.ch = ch

	return nil
}

func (c *consumer) Close() error {
	c.cancel()

	return multierr.Append(
		c.ch.Close(),
		c.conn.Close(),
	)
}
