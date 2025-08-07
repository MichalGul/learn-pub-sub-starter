package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const SimpleQueueDurable = "durable"
const SimpleQueueTransient = "transient"


// Publishes PublishJSON value of generic Type T into exchange by channel ch
// value is Marshaled to json
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	bytesVal, err := json.Marshal(val)
	if err != nil {
		fmt.Println("error marshalling data to publish", err)
		return err
	}
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        bytesVal,
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)

	return nil
}

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
		nil)
	if err != nil {
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
	handler func(T),
) error {

	channel, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return errors.New("error declaring queue " + queueName + " to exchange" + exchange)
	}

	deliveryChannel, err := channel.Consume(queueName, "", false, false, false, false, nil)


	go func() error {
		for msg := range deliveryChannel {

			var msgBody T
			marshallErr := json.Unmarshal(msg.Body, &msgBody)

			if marshallErr != nil {
				return fmt.Errorf("error unmarshalling message %v", marshallErr)
			}

			handler(msgBody)

			// acknowledge the message and remove from the queue
			msg.Ack(false)

		}
		
		return nil 
	}()


	return nil
}
