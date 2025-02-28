package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(r routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(r)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, pubChannel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(mv gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		mvOutcome := gs.HandleMove(mv)
		acktype := pubsub.NackRequeue
		switch mvOutcome {
		case gamelogic.MoveOutComeSafe:
			acktype = pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				pubChannel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: mv.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				acktype = pubsub.NackRequeue
			} else {
				acktype = pubsub.Ack
			}
		case gamelogic.MoveOutcomeSamePlayer:
			acktype = pubsub.NackDiscard
		default:
			acktype = pubsub.NackDiscard
		}
		return acktype
	}
}

func handlerWar(gs *gamelogic.GameState, pubChannel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		acktype := pubsub.NackRequeue

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			acktype = pubsub.NackRequeue
		case gamelogic.WarOutcomeOpponentWon:
			gl := routing.GameLog{
				CurrentTime: time.Now(),
				Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
				Username:    gs.GetUsername(),
			}
			if err := pubsub.PublishGob(pubChannel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), gl); err != nil {
				acktype = pubsub.NackRequeue
			} else {
				acktype = pubsub.Ack
			}

		case gamelogic.WarOutcomeYouWon:
			gl := routing.GameLog{
				CurrentTime: time.Now(),
				Message:     fmt.Sprintf("%s won a war against %s", winner, loser),
				Username:    gs.GetUsername(),
			}
			if err := pubsub.PublishGob(pubChannel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), gl); err != nil {
				acktype = pubsub.NackRequeue
			} else {
				acktype = pubsub.Ack
			}
		case gamelogic.WarOutcomeDraw:
			gl := routing.GameLog{
				CurrentTime: time.Now(),
				Message:     fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
				Username:    gs.GetUsername(),
			}
			if err := pubsub.PublishGob(pubChannel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), gl); err != nil {
				acktype = pubsub.NackRequeue
			} else {
				acktype = pubsub.Ack
			}
		default:
			fmt.Println("Unknown error happened")
			acktype = pubsub.NackDiscard
		}
		return acktype
	}
}
