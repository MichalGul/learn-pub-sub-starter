package main

import (
	"fmt"

	"log"

	"github.com/MichalGul/learn-pub-sub-starter/internal/gamelogic"
	"github.com/MichalGul/learn-pub-sub-starter/internal/pubsub"
	"github.com/MichalGul/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	rabbitConnectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitConnectionString)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer connection.Close()

	fmt.Println("Successfuly connected to RabbitMq server")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Client server failed to run: %v", err)
	}

	pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		"transient")

	gameState := gamelogic.NewGameState(userName)

	for {
		commands := gamelogic.GetInput()
		if len(commands) == 0 {
			continue
		}

		switch commands[0] {
		case "spawn":
			err := gameState.CommandSpawn(commands)
			if err != nil {
				log.Printf("could not spawn unit: %v", err)
			}

		case "move":
			armyMove, err := gameState.CommandMove(commands)
			if err != nil {
				log.Printf("could not move unit: %v", err)
			} else {
				log.Printf("Army moved to %s", armyMove.ToLocation)
			}
		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			log.Printf("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			log.Printf("Unknown command")

		}

	}

}
