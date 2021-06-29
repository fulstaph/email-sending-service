package main

import (
	"log"

	"email-sender/internal/system/applications"
)

func main() {
	acceptor, err := applications.NewAcceptor()
	if err != nil {
		log.Fatal(err)
	}

	if err = acceptor.Run(); err != nil {
		log.Fatal(err)
	}
}
