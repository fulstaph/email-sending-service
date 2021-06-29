package events

import (
	"time"
)

type Event struct {
	exchangeName string
	exchangeType string

	Meta    *Meta       `json:"meta"`
	Payload interface{} `json:"payload"`
}

type Meta struct {
	SentAt     time.Time `json:"sent_at"`
	ReceivedAt time.Time `json:"received_at"`
}

func (e *Event) Name() string {
	return e.exchangeName
}

func (e *Event) Type() string {
	return e.exchangeType
}
