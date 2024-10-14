package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DocumentChange struct {
	DocId  string `json:"doc_id"`
	Change string `json:"change"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Send(body []byte) { // Ensure that you are getting the correct environment variable
	rabbitMQDefaultUser := os.Getenv("RABBITMQ_DEFAULT_USER")
	if rabbitMQDefaultUser == "" {
		rabbitMQDefaultUser = "guest"
	}
	rabbitMQDefaultPass := os.Getenv("RABBITMQ_DEFAULT_PASS")
	if rabbitMQDefaultPass == "" {
		rabbitMQDefaultPass = "guest"
	}
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = fmt.Sprintf("amqp://%s:%s@localhost:5672/", rabbitMQDefaultUser, rabbitMQDefaultPass)
	}
	conn, err := amqp.Dial(rabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"Hello", // name
		false,   // durable
		false,   // delete when unusable
		false,   // exclusive
		false,   // no-wait
		nil,     //arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Create a new DocumentChange struct and marshall it to JSON
	docChange := DocumentChange{
		DocId:  "123",
		Change: string(body),
	}

	jsonDocChange, err := json.Marshal(docChange)
	failOnError(err, "Failed to marshal document change to JSON")

	// Set a timeout context for the publish operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish the JSON encoded message to the queue
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonDocChange,
		},
	)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", body)
}
