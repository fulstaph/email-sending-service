package main

import (
	"log"

	"email-sender/internal/system/applications"
)

func main() {
	sender, err := applications.NewSender()
	if err != nil {
		log.Fatal(err)
	}

	if err = sender.Run(); err != nil {
		log.Fatal(err)
	}
}
