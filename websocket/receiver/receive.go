package receiver

import (
	"log"

	"backend/concurrency/websocket/connect"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

// FailOnError is a helper function to log and exit on error
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Receive() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare a queue to receive messages
	q, err := ch.QueueDeclare(
		"Goodbye", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	FailOnError(err, "Failed to register a consumer")

	// Log messages as they are received
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// Lock the clients map to ensure thread safety
			connect.Mutex.Lock()

			// Broadcast the message to all clients
			for client := range connect.Clients {
				var writeErr error
				if writeErr = client.WriteMessage(websocket.TextMessage, d.Body); writeErr != nil {
					log.Println("Error writing message:", writeErr)
					break
				}
			}

			connect.Mutex.Unlock()
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
