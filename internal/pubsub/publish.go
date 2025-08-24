package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
	publishErr := ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if publishErr != nil {
		fmt.Printf("error publishing message to queue: %v\n", publishErr)
		return publishErr
	}

	return nil
}

// Publishes PublishGob value of generic Type T into exchange by channel ch
// value is parsed to gob type

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {

	var bytesBuffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&bytesBuffer)

	gobEncodingError := gobEncoder.Encode(val)
	if gobEncodingError != nil {
		fmt.Printf("error encoding data to job to publish: %v \n", gobEncodingError)
		return gobEncodingError
	}

	message := amqp.Publishing{
		ContentType: "application/gob",
		Body: bytesBuffer.Bytes(),
	}

	publishErr := ch.PublishWithContext(context.Background(), exchange, key, false, false, message)
	if publishErr != nil {
		fmt.Printf("error publishing message to queue: %v\n", publishErr)
		return publishErr
	}

	return nil
}



