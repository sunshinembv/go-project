package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	if err := nc.Publish("greetings", []byte("Hello, NATS!")); err != nil {
		log.Fatal(err)
	}

	if err := nc.Flush(); err != nil {
		log.Fatal(err)
	}
}
