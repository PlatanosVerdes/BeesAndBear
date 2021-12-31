//Consumidores de permisos
package main

import (
	"log"
	"os"
	"time"
	"strings"
	"math/rand"
	amqp "github.com/streadway/amqp"
)

const (
	maxVelocity 	 = 10 
	minVelocity 	 = 1 
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Variables y canales
	canProduce := make(chan int)
	end := make(chan int)

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
		"logs",   // name
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
		"Alert",  		// name
		false,   		// durable
		false,  		// delete when unused
		false,   		// exclusive
		false,   		// no-wait
		nil,     		// arguments
	)
	failOnError(err, "Failed to declare a queue")

	//Enlace de cola con canal
	err = channel.QueueBind(
		qAlert.Name, // queue name
		"",     	 // routing key
		"logs", 	 // exchange
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

	go func() {
		for {
			d := <- alertMsgs
			if len(d.Body) != 0{
				log.Printf("%s has given me (%s) permission to create honey", d.Body, beeName)
			}
			canProduce <- 1
		}
	}()

	permitsMsgs, err := channel.Consume(
		qPermits.Name, // queue
		beeName,       // consumer
		true,   	   // auto-ack
		false,  	   // exclusive
		false,  	   // no-local
		false,  	   // no-wait
		nil,    	   // args
	)
	failOnError(err, "Failed to register a consumer")

	go func(){
		for {
			<- canProduce
			permit := <- permitsMsgs
			//L'abella Maya produeix mel 9
			time.Sleep(time.Duration(velocity) * time.Second)
			message_body := strings.Split(string(permit.Body), " ") 
			log.Printf("La abeja " + beeName + " produce la miel " + message_body[1] + " para " + message_body[0])
			honeyValue := strings.Split(message_body[1], "/")
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

			}else{
		    	err = channel.Publish(
		    		"",     		// exchange
		    		qAlert.Name, 	// routing key
		    		false,  		// mandatory
		    		false,  		// immediate
		    		amqp.Publishing{
		    			ContentType: "text/plain",
		    			Body:        []byte(beeName),
		    		})
		    	failOnError(err, "Failed to publish a message")
			}
		}
	}()
	log.Printf("i'm %s with speed: %d", beeName, velocity)
	<- end 
}
