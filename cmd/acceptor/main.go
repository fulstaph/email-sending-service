package main

import (
	log "github.com/sirupsen/logrus"
	"projects/email-sending-service/config"
	"projects/email-sending-service/internal/broker"
	"projects/email-sending-service/internal/repository"
	"projects/email-sending-service/internal/server"
	"projects/email-sending-service/internal/service"
)

func init() {

}

func main() {
	var l = log.New()

	cfg, err := config.Load(config.DefaultPath)
	if err != nil {
		l.Fatal(err)
	}

	repo := repository.New(cfg.Database)

	err = repo.Connect()
	if err != nil {
		l.Fatal(err)
	}
	defer repo.Close()

	mq := broker.NewMessageQueue(cfg.Broker)
	err = mq.Open()
	if err != nil {
		l.Fatal(err)
	}
	defer mq.Close()

	acceptor := service.NewAcceptor(repo, mq)

	app := server.New(acceptor, cfg.Server, l)

	app.Start()
}
