package events

import (
	"time"

	"email-sender/internal/entities"

	"github.com/streadway/amqp"
)

func NewNotificationCreatedEvent(payload *entities.Notification, exchangeName string) (*Event, error) {
	return &Event{
		exchangeName: exchangeName,
		exchangeType: amqp.ExchangeFanout,
		Meta: &Meta{
			SentAt: time.Now(),
		},
		Payload: payload,
	}, nil
}
