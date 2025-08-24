package main

import (
	"fmt"
	"log"

	"github.com/MichalGul/learn-pub-sub-starter/internal/gamelogic"
	"github.com/MichalGul/learn-pub-sub-starter/internal/pubsub"
	"github.com/MichalGul/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)


func handlerGameLogPassed() func(routing.GameLog) pubsub.Acktype {

	return func(gl routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")
		gamelogic.WriteLog(gl)
		return pubsub.Ack
	}

}


func main() {
	fmt.Println("Starting Peril server...")

	rabbitConnectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitConnectionString)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer connection.Close()

	fmt.Println("Peril game server successfuly connected to RabbitMq server")

	//Creating new channel
	mainChannel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error creating to Channel: %v", err)
	}

	// Declare and bind queue to peril_topic
	_, queue, err:= pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
	)

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerGameLogPassed(),
	)




	gamelogic.PrintClientHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			log.Println("Sending pause message")
			err := pubsub.PublishJSON(mainChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Printf("could not publish pause: %v", err)
			}

		case "resume":
			log.Println("Sending resume message")
			err := pubsub.PublishJSON(mainChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Printf("could not publish resume: %v", err)
			}

		case "quit":
			log.Println("Exiting game")
			return

		default:
			log.Println("Unknown command.")

		}

	}

	// wait for ctrl+c
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("RabbitMQ connection closed.")

}
