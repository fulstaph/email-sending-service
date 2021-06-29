package producer

import (
	"encoding/json"
	"fmt"
	"time"

	"email-sender/internal/system/broker/events"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

const defaultRoutingKey = ""

type Producer interface {
	Produce(event *events.Event) error
}

type producer struct {
	client Client
}

func New(client Client) Producer {
	return &producer{
		client: client,
	}
}

func (p *producer) Produce(event *events.Event) error {
	retryStrategy := backoff.NewExponentialBackOff()
	for {
		err := p.produce(event)
		if err == nil {
			return nil
		}

		backOff := retryStrategy.NextBackOff()
		if backOff < 1 {
			return err
		}

		time.Sleep(backOff)
	}
}

func (p *producer) produce(event *events.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to prepare event to publish: %w", err)
	}

	err = p.client.Channel().ExchangeDeclare(
		event.Name(),
		event.Type(),
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	err = p.client.Channel().Publish(
		event.Name(),
		defaultRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}
