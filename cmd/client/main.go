package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connectString = "amqp://guest:guest@localhost:5672/"
const deadLetterExchangeName = "peril_dlx"

func main() {
	fmt.Println("Starting Peril client...")
	connec, err := amqp.Dial(connectString)
	if err != nil {
		fmt.Printf("couldn't connect to RabbitMQ: %v", err)
		return
	}
	defer connec.Close()

	chanl, err := connec.Channel()
	if err != nil {
		fmt.Printf("couldn't open channel: %v", err)
		return
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("couldn't welcome client: %v", err)
		return
	}

	exchange := routing.ExchangePerilDirect
	key := routing.PauseKey
	queueName := key + "." + username
	simpleQueueType := 1
	args := amqp.Table{"x-dead-letter-exchange": deadLetterExchangeName}

	// _, _, err = pubsub.DeclareAndBind(connec, exchange, queueName, key, simpleQueueType, args)
	// if err != nil {
	// 	fmt.Printf("couldn't daclare and bind a queue: %v", err)
	// 	return
	// }

	moveExchange := routing.ExchangePerilTopic
	moveKey := routing.ArmyMovesPrefix + ".*"
	moveQueueName := routing.ArmyMovesPrefix + "." + username
	movePubKey := routing.ArmyMovesPrefix + "." + username
	// _, _, err = pubsub.DeclareAndBind(connec, moveExchange, moveQueueName, moveKey, 1, args)
	// if err != nil {
	// 	fmt.Printf("couldn't daclare and bind a queue: %v", err)
	// 	return
	// }

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(connec, exchange, queueName, key, simpleQueueType, handlerPause(gameState), args)
	if err != nil {
		fmt.Printf("couldn't subscribe JSON: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(connec, moveExchange, moveQueueName, moveKey, 1, handlerMove(gameState, chanl), args)
	if err != nil {
		fmt.Printf("couldn't subscribe JSON: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(connec, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", 0, handlerWar(gameState, chanl), args)
	if err != nil {
		fmt.Printf("couldn't subscribe JSON: %v", err)
		return
	}

gameloop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			if err := gameState.CommandSpawn(words); err != nil {
				fmt.Printf("error spawning unit: %v", err)
				continue
			}
		case "move":
			mv, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("error moving unit: %v", err)
				continue
			}
			if err = pubsub.PublishJSON(chanl, moveExchange, movePubKey, mv); err != nil {
				fmt.Printf("error publishing move message: %v", err)
				continue
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(words) != 2 {
				fmt.Println("Invalid input, need spam {number}")
				continue
			}
			num, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("fail to convert string to integer")
				continue
			}
			for i := 0; i < num; i++ {
				mlog := gamelogic.GetMaliciousLog()
				if err = pubsub.PublishGob(
					chanl,
					routing.ExchangePerilTopic,
					routing.GameLogSlug+"."+gameState.GetUsername(),
					routing.GameLog{
						CurrentTime: time.Now(),
						Message:     mlog,
						Username:    gameState.GetUsername(),
					}); err != nil {
					fmt.Println("fail to publish malicious message")
					continue
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			break gameloop
		default:
			fmt.Println("Invalid command!")
		}
	}

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("Terminate the program, connection is closed")

}
