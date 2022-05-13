package client

import (
	"context"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/renicer"
	"github.com/blocktop/pocket-autonice/zeromq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	subscriber *zeromq.Subscriber
)

func Start(ctx context.Context) {
	topics := []string{"ping"}
	chains := viper.GetStringMapString(config.Chains)
	for chainID := range chains {
		topics = append(topics, chainID)
	}

	messageChan := make(chan string, 256)

	subscriber = zeromq.NewSubscriber(topics, messageChan)
	defer subscriber.Close()
	if err := subscriber.Start(); err != nil {
		log.Fatalf(err.Error())
	}

	log.Infof("starting message consumer on %s", viper.GetString(config.SubscriberBindAddress))

	go processMessages(ctx, messageChan)

	<-ctx.Done()

	log.Info("stopping message consumer")
}

func processMessages(ctx context.Context, messageChan chan string) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("exiting client loop")
			return
		case msg := <-messageChan:
			log.Debugf("consumer received message %s", msg)
			processMessage(ctx, msg)
		}
	}
}

func processMessage(ctx context.Context, msg string) {
	if string(msg) == "ping" {
		log.Info("consumer received ping")
		return
	}
	renicer.Renice(ctx, msg)
}
