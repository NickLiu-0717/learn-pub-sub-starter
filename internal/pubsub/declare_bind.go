package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

const (
	DurableQueue   = iota // will be 0
	TransientQueue        // will be 1
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	args amqp.Table,
) (*amqp.Channel, amqp.Queue, error) {

	chanl, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	durable := simpleQueueType == DurableQueue
	autoDelete := simpleQueueType == TransientQueue
	exclusive := simpleQueueType == TransientQueue
	queue, err := chanl.QueueDeclare(queueName, durable, autoDelete, exclusive, false, args)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	if err = chanl.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}
	return chanl, queue, nil
}
