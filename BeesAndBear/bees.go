// Para mas información, visita el GitHub: https://github.com/JorgeGonzalezPascual/BeesAndBear
// Autores: Jorge González Pascual, Rubén Palmer Pérez
// Video : https://youtu.be/kbNPbpfEDM4

package main

import (
	"log"							 // Prints
	"os"							 // Get arguments
	"time"							 // Sleep
	"strings"						 // Split strings
	"math/rand"						 // Create random numbers

	amqp "github.com/streadway/amqp" // RabbitMQ
)

const (
	maxVelocity  = 10 
	minVelocity  = 6  // >= bear.speed * potsize para un mejor output 
	bearColor    = "\033[36m"
	beeColor	 = "\033[33m"
	reset 		 = "\033[0m"
)

func printBee(input string)string{
	return string(beeColor) + input + string(reset)
}

func printBear(input string) string{
	return string(bearColor) + input + string(reset)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Generate random velocity for each bee
	rand.Seed(time.Now().UnixNano())
	velocity := minVelocity + rand.Intn(maxVelocity-minVelocity)

	// Asignacion de nombre
	beeName := "Bee"
	
	if (len(os.Args) != 1) {
		beeName = os.Args[1]
	}
	// Inicializar canales en Rabbit Mq

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
	
	// Cola de Alerta a las abejas de tarro lleno
	qAlert, err := channel.QueueDeclare(
		"",  			// name
		false,   		// durable
		false,  		// delete when unused
		true,   		// exclusive
		false,   		// no-wait
		nil,     		// arguments
	)
	failOnError(err, "Failed to declare a queue")

	//Enlace de cola con canal
	err = channel.QueueBind(
		qAlert.Name, // queue name
		"",     	 // routing key
		"bees", 	 // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")
	
	alertMsgs, err := channel.Consume(
		qAlert.Name, // queue
		"",     // consumer
		true,   	 // auto-ack
		false,  	 // exclusive
		false,  	 // no-local
		false,  	 // no-wait
		nil,    	 // args
	)
	failOnError(err, "Failed to register a consumer")

	permitsMsgs, err := channel.Consume(
		qPermits.Name, 	// queue
		"",       		// consumer
		false,   	   	// auto-ack
		false,  	   	// exclusive
		false,  	   	// no-local
		false,  	   	// no-wait
		nil,    	   	// args
	)
	failOnError(err, "Failed to register a consumer")

	go func(){
		for {
			
			permit := <- permitsMsgs
		    if len(permit.Body) != 0{ // Fix 
				
				//L'abella Maya produeix mel 9
				message_body := strings.Split(string(permit.Body), " ") 
				honeyValue := strings.Split(message_body[1], "/")
				// Producir Miel
				time.Sleep(time.Duration(velocity) * time.Second)
				log.Printf(printBee(beeName) + ": produce la miel " + message_body[1] + " para " + printBear(message_body[0]))

				// If maxSize/maxSize -> we're done
				if honeyValue[0] == honeyValue[1]{
		    		err = channel.Publish(
		    			"",     		// exchange
		    			qWakeUp.Name, 	// routing key
		    			false,  		// mandatory
		    			false,  		// immediate
		    			amqp.Publishing{
		    				ContentType: "text/plain",
		    				Body:        []byte(beeName),
		    			})
		    		failOnError(err, "Failed to publish a message")
				}
			}
			permit.Ack(false)
		}
	}()
	log.Printf("Esta es la abeja %s con velocidad: %d", printBee(beeName), velocity)
	<-alertMsgs 
	log.Printf("%s: El bote esta roto y me voy", printBee(beeName))
}