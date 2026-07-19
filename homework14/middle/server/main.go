package main

import (
	"fmt"
	"log"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	msg, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	var result int
	for data := range msg {
		n, err := strconv.Atoi(string(data.Body))
		if err != nil {
			log.Printf("invalid request %q: %v", string(data.Body), err)

			if err := data.Reject(false); err != nil {
				log.Printf("reject invalid request: %v", err)
			}

			continue
		}

		fmt.Printf("Server: сonsume num = %d\n", n)
		result = 2 * n

		err = ch.Publish(
			"",
			data.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: data.CorrelationId,
				Body:          []byte(strconv.Itoa(result)),
			},
		)
		if err != nil {
			log.Printf("publish response: %v", err)

			if err := data.Nack(false, true); err != nil {
				log.Printf("nack request: %v", err)
			}

			continue
		}

		if err := data.Ack(false); err != nil {
			log.Printf("ack request: %v", err)
		}
	}
}
