package config

import (
	"encoding/json"
	"io/ioutil"
)

var DefaultPath = "config/config.json"

type Config struct {
	Database      Database      `json:"database"`
	Server        Server        `json:"server"`
	Broker        Broker        `json:"broker"`
	MailingServer MailingServer `json:"mailing_server"`
}

type Database struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Name       string `json:"name"`
	Collection string `json:"collection"`
}

type Server struct {
	Port string `json:"port"`
}

type Broker struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	QueueName  string `json:"queue_name"`
	Exchange   string `json:"exchange"`
	RoutingKey string `json:"routing_key"`
}

type MailingServer struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

func Load(path string) (*Config, error) {
	fileBody, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(fileBody, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
