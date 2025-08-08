package main

import (
	"fmt"
	"log"

	"github.com/MichalGul/learn-pub-sub-starter/internal/gamelogic"
	"github.com/MichalGul/learn-pub-sub-starter/internal/pubsub"
	"github.com/MichalGul/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

// HANDLERS FOR PUBLISHED MOVES
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {

	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}

}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {

	return func(mv gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(mv)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}


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

	// Declare for direct exchange for pause messages
	channel, _, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		"transient")

	gameState := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, userName),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)

	if err != nil {
		log.Fatalf("Error subscribing to Direct exchange pause queue: %v", err)
	}

	//Subscribe to moves from other players exchange army_moves.*
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+userName,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

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
				// publish move message to all subscribents
				err := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+userName, armyMove)
				if err != nil {
					log.Printf("publishing move failed: %v", err)
				} else {
					log.Printf("Published message: Army with units %s moved to %s", len(armyMove.Units), armyMove.ToLocation)
				}
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
