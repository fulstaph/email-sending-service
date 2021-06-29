package config

type Consumer struct {
	ConnectionURL      string
	NotificationsQueue *Queue
}

type Queue struct {
	Name            string
	Durable         bool   `envconfig:"default=true,optional"`
	RoutingKey      string `envconfig:"optional"`
	Exchange        string
	ExchangeDurable bool `envconfig:"default=true,optional"`
}
