//Consumidores de permisos
package main

import (
	"log"

	amqp "github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable (the queue will survive a broker restart)
		false,   // delete when unused
		false,   // exclusive (used by only one connection and the queue will be deleted when that connection closes)
		false,   // no-wait (the server will not respond to the method. The client should not wait for a reply method)
		nil,     // arguments (Those are provided by clients when they declare queues (exchanges) and control various optional features, such as queue length limit or TTL.)
	)

	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume( //This method asks the server to start a "consumer",
		// which is a transient request for messages from a specific queue.
		// Consumers last as long as the channel they were declared on,
		// or until the client cancels them.
		q.Name, // queue
		"",     // consumer (Specifies the identifier for the consumer. The consumer tag is local to a channel, so two clients can use the same consumer tags. If this field is empty the server will generate a unique tag.)
		true,   // auto-ack (If this field is set the server does not expect acknowledgements for messages. That is, when a message is delivered to the client the server assumes the delivery will succeed and immediately dequeues it. This functionality may increase performance but at the cost of reliability. Messages can get lost if a client dies before they are delivered to the application.)
		false,  // exclusive (Request exclusive consumer access, meaning only this consumer can access the queue.)
		false,  // no-local (If the no-local field is set the server will not send messages to the connection that published them.)
		false,  // no-wait (If set, the server will not respond to the method. The client should not wait for a reply method. If the server could not complete the method it will raise a channel or connection exception.)
		nil,    // args (A set of arguments for the consume. The syntax and semantics of these arguments depends on the server implementation.)
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}