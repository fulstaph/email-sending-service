package main

import (
	"log"
	"projects/email-sending-service/config"
	"projects/email-sending-service/internal/repository"
	"projects/email-sending-service/internal/server"
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

	acceptor := service.NewAcceptor(repo)

	app := server.New(acceptor, cfg.Server)

	app.Start()
}
