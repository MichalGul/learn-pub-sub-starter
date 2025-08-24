package pubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string
type Acktype int

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

const SimpleQueueDurable = "durable"
const SimpleQueueTransient = "transient"


// Declares and binds a queue
func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	//Creating new channel
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, errors.New("Error creating Channel")
	}

	declaredQueue, err := channel.QueueDeclare(queueName,
		queueType == "durable",
		queueType == "transient",
		queueType == "transient",
		false,
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"})

	if err != nil {
		log.Println(err)
		return channel, amqp.Queue{}, errors.New("Error declaring queue")
	}

	errBind := channel.QueueBind(queueName, key, exchange, false, nil)
	if errBind != nil {
		errorMsg := fmt.Sprintf("Error binding queue %s to exchange %s", queueName, exchange)
		return channel, amqp.Queue{}, errors.New(errorMsg)
	}

	return channel, declaredQueue, nil

}

// Subscribe to Queue
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {

	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return errors.New("error declaring queue " + queueName + " to exchange" + exchange)
	}

	deliveryChannel, err := channel.Consume(queueName, "", false, false, false, false, nil)

	go func() error {
		defer channel.Close()
		for msg := range deliveryChannel {

			var msgBody T
			marshallErr := json.Unmarshal(msg.Body, &msgBody)

			if marshallErr != nil {
				return fmt.Errorf("error unmarshalling message %v", marshallErr)
			}

			messageAckinfo := handler(msgBody)

			// acknowledge the message and remove from the queue
			switch messageAckinfo {
			case Ack:
				msg.Ack(false)
				log.Println("Message was acknowledged: Ack")
			case NackRequeue:
				msg.Nack(false, true)
				log.Println("Message was not acknowledged: NackRequeue")
			case NackDiscard:
				msg.Nack(false, false)
				log.Println("Message was not acknowledged: NackDiscard")
			}
		}

		return nil
	}()

	return nil
}
