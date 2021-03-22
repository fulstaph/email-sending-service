package main

import (
	log "github.com/sirupsen/logrus"
	"projects/email-sending-service/config"
	"projects/email-sending-service/internal/broker"
	"projects/email-sending-service/internal/repository"
	"projects/email-sending-service/internal/service"
)

func main() {
	cfg, err := config.Load(config.DefaultPath)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.New(cfg.Database)

	err = repo.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	mq := broker.NewMessageQueue(cfg.Broker)
	err = mq.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	sender := service.NewSender(
		repo,
		mq,
		cfg.MailingServer,
	)

	if err := sender.Start(); err != nil {
		log.Fatal(err)
	}

	// kool trick
	forever := make(chan bool)
	<-forever
}
