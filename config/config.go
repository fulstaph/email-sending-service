package config

import (
	"encoding/json"
	"io/ioutil"
)

var DefaultPath = "config/config.json"

type Config struct {
	Database Database `json:"database"`
	Server   Server   `json:"server"`
	Broker   Broker   `json:"broker"`
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
