package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: run main.go <number>")
	}

	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("invalid number %q: %v", os.Args[1], err)
	}

	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to close resource: %v\n", err)
		}
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			fmt.Printf("Failed to close resource: %v\n", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	correlationID := uuid.NewString()

	err = ch.PublishWithContext(
		ctx,
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: correlationID,
			ReplyTo:       q.Name,
			Body:          []byte(strconv.Itoa(n)),
		},
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

	for {
		select {
		case <-ctx.Done():
			log.Printf("Timeout: %v", ctx.Err())
			return
		case data, ok := <-msg:
			if !ok {
				log.Println("reply channel closed")
				return
			}

			if data.CorrelationId != correlationID {
				if err := data.Reject(false); err != nil {
					log.Printf("reject unexpected response: %v", err)
					return
				}
				continue
			}

			result, err := strconv.Atoi(string(data.Body))
			if err != nil {
				log.Printf("invalid server response: %v", err)
				return
			}

			fmt.Printf("Result: %d\n", result)

			if err := data.Ack(false); err != nil {
				log.Printf("ack response: %v", err)
			}

			return
		}
	}
}
