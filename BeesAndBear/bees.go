//Consumidores de permisos
package main

import (
	"log"
	"os"
	amqp "github.com/streadway/amqp"
)

const (
	velocity 	 = 1
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func initRabbit(){
	// Conexion
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	
	// Canal Oso -> Abeja (poner permisos para abeja)
	chPermits, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer chPermits.Close()

	// Cola
	qPermits, err := chPermits.QueueDeclare(
		"Permisos", // name
		false,      // durable
		false,   	// delete when unused
		false,  	// exclusive
		false, 		// no-wait
		nil,   		// arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Canal Abeja -> Oso (despertar oso)
	chWakeUp, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer chWakeUp.Close()

	// Cola
	qWakeUp, err := chWakeUp.QueueDeclare(
		"Wake Up",  // name
		false,   	// durable
		false,   	// delete when unused
		false,   	// exclusive
		false,   	// no-wait
		nil,     	// arguments
	)
	failOnError(err, "Failed to declare a queue")
	
	// Canal Oso -> Abeja (avisar Abejas)
	chAlert, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer chAlert.Close()
	
	// Cola
	qAlert, err := chAlert.QueueDeclare(
		"Alert",  		// name
		false,   		// durable
		false,  		// delete when unused
		false,   		// exclusive
		false,   		// no-wait
		nil,     		// arguments
	)
	failOnError(err, "Failed to declare a queue")
}

func main() {
	//Asignacion de nombre
	beeName := "Bee"
	
	if (len(os.Args) != 1) {
		beeName = os.Args[1]
	}
	// Inicializar canales en Rabbit Mq

	// Conexion
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	
	// Canal Oso -> Abeja (poner permisos para abeja)
	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	// Cola Permisos
	qPermits, err := channel.QueueDeclare(
		"Permisos", // name
		false,      // durable
		false,   	// delete when unused
		false,  	// exclusive
		false, 		// no-wait
		nil,   		// arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Cola Despertador
	qWakeUp, err := channel.QueueDeclare(
		"Wake Up",  // name
		false,   	// durable
		false,   	// delete when unused
		false,   	// exclusive
		false,   	// no-wait
		nil,     	// arguments
	)
	failOnError(err, "Failed to declare a queue")
	
	// Cola de Alerta
	qAlert, err := channel.QueueDeclare(
		"Alert",  	// name
		false,   	// durable
		false,  	// delete when unused
		false,   	// exclusive
		false,   	// no-wait
		nil,     	// arguments
	)
	failOnError(err, "Failed to declare a queue")

	//forever := make(chan bool)

	msgs, err := channel.Consume(
		qAlert.Name, // queue
		beeName,     	 // consumer
		true,   	 // auto-ack
		false,  	 // exclusive
		false,  	 // no-local
		false,  	 // no-wait
		nil,    	 // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			log.Printf("%s produce: %s", beeName, d.Body)
		}
	}()
	

	log.Printf(" [*] " + beeName + " Waiting for messages. To exit press CTRL+C")
	//<-forever

}
