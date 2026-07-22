package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("greetings", func(msg *nats.Msg) {
		fmt.Println(string(msg.Data))
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := nc.Flush(); err != nil {
		log.Fatal(err)
	}

	select {}
}
