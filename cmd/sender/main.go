package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"projects/email-sending-service/internal/broker"
	"time"
)

func main() {
	conn, err := broker.GetConn("amqp://guest:guest@localhost")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(time.Second)
			conn.Publish("test-key", []byte(`{"message":"test"}`))
		}
	}()

	err = conn.Consumer("test-queue", "test-key", handler, 2)

	if err != nil {
		log.Fatal(err)
	}

	forever := make(chan bool)
	<-forever
}

func handler(d amqp.Delivery) bool {
	if d.Body == nil {
		fmt.Println("Error, no message body!")
		return false
	}
	fmt.Println(string(d.Body))
	return true
}