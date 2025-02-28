package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) Acktype,
	args amqp.Table,
) error {
	return subscribe(conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		JSONUnmarshaller,
		args)
}

func SubscribeGOB[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) Acktype,
	args amqp.Table,
) error {
	return subscribe(conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		GOBUnmarshaller,
		args)

}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
	args amqp.Table,
) error {

	_, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType, args)
	if err != nil {
		return err
	}

	chanl, err := conn.Channel()
	if err != nil {
		return err
	}
	err = chanl.Qos(10, 0, false)
	if err != nil {
		return err
	}
	delivery, err := chanl.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		defer chanl.Close()
		for de := range delivery {
			dat, err := unmarshaller(de.Body)
			if err != nil {
				log.Printf("couldn't unmarshal body: %v", err)
				continue
			}
			whatack := handler(dat)
			switch whatack {
			case Ack:
				if err = de.Ack(false); err != nil {
					log.Printf("couldn't Ack the message: %v", err)
					continue
				}
			case NackRequeue:
				if err = de.Nack(false, true); err != nil {
					log.Printf("couldn't Nack Requeue the message: %v", err)
					continue
				}
			case NackDiscard:
				if err = de.Nack(false, false); err != nil {
					log.Printf("couldn't Nack Discard the message: %v", err)
					continue
				}
			}

		}
	}()
	return nil
}

func JSONUnmarshaller[T any](data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}

func GOBUnmarshaller[T any](data []byte) (T, error) {
	var v T
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&v)
	return v, err
}
