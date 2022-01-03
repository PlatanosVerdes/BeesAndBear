// Para mas información, visita el GitHub: https://github.com/JorgeGonzalezPascual/BeesAndBear
// Autores: Jorge González Pascual, Rubén Palmer Pérez
// Video : https://youtu.be/kbNPbpfEDM4

package main

import (
	"log"							 // Prints
	"time"							 // Sleep
	"os"							 // Get arguments
	"strconv" 						 // Convert strings to int and viceversa

	amqp "github.com/streadway/amqp" // RabbitMQ
)

const (
	potSize     = 5
	timesToEat 	= 2
	velocity 	= 1
	bearColor   = "\033[36m"
	beeColor	= "\033[33m"
	reset 		= "\033[0m"
)

// Output coloreado 
func printBear(input string) string{
	return string(bearColor) + input + string(reset)
}

func printBee(input string)string{
	return string(beeColor) + input + string(reset)
}

// Funcion de error Rabbit Mq
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	// variables y canales
	var pot [potSize]string

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
	
	// Canal Oso - Abeja
	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	// Fanout del canal
	err = channel.ExchangeDeclare(
		"bees",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Cola Permisos para las abejas
	qPermits, err := channel.QueueDeclare(
		"Permisos", // name
		false,      // durable
		false,   	// delete when unused
		false,  	// exclusive
		false, 		// no-wait
		nil,   		// arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Cola Despertador del oso
	qWakeUp, err := channel.QueueDeclare(
		"Wake Up",  // name
		false,   	// durable
		false,   	// delete when unused
		false,   	// exclusive
		false,   	// no-wait
		nil,     	// arguments
	)
	failOnError(err, "Failed to declare a queue")
	

	msgsWakeUp, err := channel.Consume(
		qWakeUp.Name, // queue
		bearName,     // consumer
		true,   	  // auto-ack
		false,  	  // exclusive
		false,  	  // no-local
		false,  	  // no-wait
		nil,    	  // args
	)
	failOnError(err, "Failed to register a consumer")
	// END RabbitMq init

	for i := 0; i < timesToEat; i++ { // Times to eat
		for j := 0; j < potSize; j++ { // Size of pot

			// Produce
			time.Sleep(velocity * time.Second)
			pot[j] = bearName + " " +  strconv.Itoa(j+1) + "/" + strconv.Itoa(potSize) // nombreOso x/potSize (ej: Baloo 1/5)
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

		log.Printf(printBear(bearName) + " se va a dormir")

		// Se va a dormir 
		beeName := <- msgsWakeUp

		log.Printf("A %s le ha levantado %s y come %d/%d",printBear(bearName), printBee(string(beeName.Body)), i+1, timesToEat)
	}
	log.Printf("%s esta lleno y rompe el bote de miel !!!", printBear(bearName))

	err = channel.Publish(
		"bees",     		// exchange
		"", 	// routing key
		false,  		// mandatory
		false,  		// immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("exit"),
		})
	failOnError(err, "Failed to publish a message")
}