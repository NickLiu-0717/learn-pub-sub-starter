package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	mesg, err := json.Marshal(val)
	if err != nil {
		return err
	}

	pub := amqp.Publishing{
		ContentType: "application/json",
		Body:        mesg,
	}

	if err = ch.PublishWithContext(context.Background(), exchange, key, false, false, pub); err != nil {
		return err
	}
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var net bytes.Buffer
	if err := gob.NewEncoder(&net).Encode(val); err != nil {
		return err
	}

	pub := amqp.Publishing{
		ContentType: "application/gob",
		Body:        net.Bytes(),
	}

	if err := ch.PublishWithContext(context.Background(), exchange, key, false, false, pub); err != nil {
		return err
	}
	return nil
}
