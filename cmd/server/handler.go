package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(wl routing.GameLog) pubsub.Acktype {
	return func(wl routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")
		if err := gamelogic.WriteLog(wl); err != nil {
			fmt.Printf("couldn't write log: %v", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
