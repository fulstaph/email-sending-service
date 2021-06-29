package config

import (
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type ConfigSender struct {
	LogLevel    string
	AppName     string
	MetricsPort string
	Port        string
	SMTP        *SMTP
	Database    *Database
	Consumer    *Consumer
}

type ConfigAcceptor struct {
	LogLevel    string
	AppName     string
	MetricsPort string
	Port        string
	Database    *Database
	Producer    *Producer
}

func LoadSender() (*ConfigSender, error) {
	var cfg ConfigSender
	if err := envconfig.Init(&cfg); err != nil {
		return nil, errors.Wrap(err, "error loading sender configuration")
	}

	return &cfg, nil
}

func LoadAcceptor() (*ConfigAcceptor, error) {
	var cfg ConfigAcceptor
	if err := envconfig.Init(&cfg); err != nil {
		return nil, errors.Wrap(err, "error loading acceptor configuration")
	}

	return &cfg, nil
}
