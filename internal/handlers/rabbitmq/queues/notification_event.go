package queues

import (
	"email-sender/internal/entities"
	"email-sender/internal/system/broker/events"
)

type NotificationEvent struct {
	Meta    *events.Meta           `json:"meta"`
	Payload *entities.Notification `json:"payload"`
}
