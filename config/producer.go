package config

import (
	"time"
)

type Producer struct {
	URL             string
	RetryTimeout    time.Duration
	Exchange        string
	ExchangeDurable bool `envconfig:"default=true,optional"`
}
