package client

import (
	"context"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/messaging"
	"github.com/blocktop/pocket-autonice/renicer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	subscriber *messaging.Subscriber
)

func Start(ctx context.Context) {
	topics := []string{"ping"}
	chains := viper.GetStringMapString(config.Chains)
	for chainID := range chains {
		topics = append(topics, chainID)
	}

	messageChan := make(chan messaging.PubSubMessage, 256)

	subscriber = messaging.NewSubscriber(topics, messageChan)
	defer subscriber.Close()

	log.Infof("message consumer dialing %s", viper.GetString(config.SubscriberPublisherAddress))

	if err := subscriber.Start(ctx); err != nil {
		log.Fatalf(err.Error())
	}

	go processMessages(ctx, messageChan)

	<-ctx.Done()

	log.Info("stopping message consumer")
}

func processMessages(ctx context.Context, messageChan chan messaging.PubSubMessage) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("exiting client loop")
			return
		case message := <-messageChan:
			processMessage(ctx, message)
		}
	}
}

func processMessage(ctx context.Context, message messaging.PubSubMessage) {
	if message.Topic() == "ping" {
		log.Info("consumer received ping")
		return
	}
	renicer.Renice(ctx, message.Message())
}
