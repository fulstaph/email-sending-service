package producer

import (
	"errors"
	"fmt"
	"time"

	"email-sender/config"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Client interface {
	Channel() *amqp.Channel
	ReconnectHandler()
	Close() error
}

type client struct {
	logger  *zap.Logger
	cfg     *config.Producer
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewClient(cfg *config.Producer, logger *zap.Logger) (Client, error) {
	conn, channel, err := setupConnection(cfg.URL)
	if err != nil {
		return nil, err
	}

	c := &client{
		logger:  logger,
		cfg:     cfg,
		conn:    conn,
		channel: channel,
	}

	return c, nil
}

func (c *client) ReconnectHandler() {
	go func() {
		closeErr := <-c.channel.NotifyClose(make(chan *amqp.Error))
		if closeErr == nil {
			c.logger.Info("graceful close")
			return
		}

		linearRetry := backoff.NewConstantBackOff(c.cfg.RetryTimeout)

		for {
			conn, channel, err := setupConnection(c.cfg.URL)
			if err != nil {
				c.logger.Error("failed to reconnect")
				time.Sleep(linearRetry.NextBackOff())
				continue
			}

			c.conn = conn
			c.channel = channel

			c.logger.Info("connect successful")
			break
		}

		c.ReconnectHandler()
	}()
}

func (c *client) Channel() *amqp.Channel {
	return c.channel
}

func (c *client) Close() error {
	err := multierr.Append(c.channel.Close(), c.conn.Close())
	if errors.Is(err, amqp.ErrClosed) {
		return nil
	}

	return err
}

func setupConnection(connectionURL string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial rabbitmq connect: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init rabbitmq channel: %w", err)
	}

	return conn, channel, nil
}
