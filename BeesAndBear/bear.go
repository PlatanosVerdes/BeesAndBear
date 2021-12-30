//Productor de permisos
package main

import (
	"log"
	"time"
	"os"
	"strconv"

	amqp "github.com/streadway/amqp"
)

const (
	potSize     = 5 
	timesToEat 	= 2
	velocity 	= 1
)


// Funcion de error Rabbit Mq
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}


func main() {

	// variables y canales
	chanWakeUp := make(chan int)
	asleep := make(chan int)

	// Asignacion de nombre
	bearName := "Bear"
	
	if (len(os.Args) != 1) {
		bearName = os.Args[1]
	}

	// BEGIN RabbitMq init

	// Conexion
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	
	// Canal Oso -> Abeja (poner permisos para abeja)
	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

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
		"Alert",  		// name
		false,   		// durable
		false,  		// delete when unused
		false,   		// exclusive
		false,   		// no-wait
		nil,     		// arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		qAlert.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := channel.Consume(
	qWakeUp.Name, // queue
	bearName,     // consumer
	true,   	  // auto-ack
	false,  	  // exclusive
	false,  	  // no-local
	false,  	  // no-wait
	nil,    	  // args
	)
	// END RabbitMq init

	go func(){
		for{
			<- asleep
			d := <- msgs

			// L'ha despertat l'abella: Maya i menja 2/3
			if len(d.Body) != 0{
				log.Printf("%s is woken up to: %s",bearName, d.Body)
			}
			chanWakeUp <- 1
		}
	}()

	var pot [potSize]string

	for i := 0; i < timesToEat; i++ { // Times to eat
		log.Printf(bearName + " wants the pot filled")
		for j := 0; j < potSize; j++ { // Size of pot

			// Produce
			time.Sleep(velocity * time.Second)
			pot[j] = bearName + " " +  strconv.Itoa(j+1)

			//Permisos
			err = channel.Publish(
				"",     		// exchange
				qPermits.Name,  // routing key
				false,  		// mandatory
				false,  		// immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(pot[j]),
				})
			failOnError(err, "Failed to publish a message")
		}

		//Avisar
		var message = "lets go"
		err = channel.Publish(
			"",     		// exchange
			qAlert.Name, 	// routing key
			false,  		// mandatory
			false,  		// immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
			})
		failOnError(err, "Failed to publish a message")

		log.Printf(bearName + " goes to sleep")
		asleep <- 1
		// Esperar hasta que me levanten  
	
		// Barrier, wait to get a message :D
		<- chanWakeUp
		failOnError(err, "Failed to register a consumer")
	}
	log.Printf("He salido")

	// Purge
	channel.Close()
}
