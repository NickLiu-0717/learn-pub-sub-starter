package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connectString = "amqp://guest:guest@localhost:5672/"
const deadLetterExchangeName = "peril_dlx"

func main() {
	fmt.Println("Starting Peril server...")
	connect, err := amqp.Dial(connectString)
	if err != nil {
		fmt.Printf("couldn't connect to RabbitMQ: %v", err)
		return
	}
	fmt.Println("Successfully connect to RabbitMQ")
	defer connect.Close()

	chanl, err := connect.Channel()
	if err != nil {
		fmt.Printf("couldn't open channel: %v", err)
		return
	}

	exchange := routing.ExchangePerilDirect
	key := routing.PauseKey
	pause := routing.PlayingState{
		IsPaused: true,
	}
	resume := routing.PlayingState{
		IsPaused: false,
	}
	args := amqp.Table{"x-dead-letter-exchange": deadLetterExchangeName}

	topicKey := routing.GameLogSlug + ".*"
	// _, _, err = pubsub.DeclareAndBind(connect, routing.ExchangePerilTopic, routing.GameLogSlug, topicKey, 0, args)
	// if err != nil {
	// 	fmt.Printf("couldn't daclare and bind a queue: %v", err)
	// 	return
	// }
	if err = pubsub.SubscribeGOB(connect, routing.ExchangePerilTopic, routing.GameLogSlug, topicKey, 0, handlerLog(), args); err != nil {
		fmt.Printf("couldn't subscribe JSON: %v", err)
	}

	gamelogic.PrintServerHelp()

gameloop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message")
			if err = pubsub.PublishJSON(chanl, exchange, key, pause); err != nil {
				fmt.Printf("couldn't publish json: %v", err)
				continue
			}
		case "resume":
			fmt.Println("Sending resume message")
			if err = pubsub.PublishJSON(chanl, exchange, key, resume); err != nil {
				fmt.Printf("couldn't publish json: %v", err)
				continue
			}
		case "quit":
			fmt.Println("Exiting the Peril game")
			break gameloop
		default:
			fmt.Println("Invalid command!")
		}

	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Terminate the program, connection is closed")
}
