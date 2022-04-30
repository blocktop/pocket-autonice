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
	pubsubTopic := viper.GetString(config.PubSubTopic)
	messageChan := make(chan []byte, 256)
	subscriber = zeromq.NewSubscriber(pubsubTopic, messageChan)
	defer subscriber.Close()

	log.Infof("starting message consumer on %s", viper.GetString(config.ZeroMQAddress))

	go processMessages(ctx, messageChan)

	<-ctx.Done()

	log.Info("stopping message consumer")
}

func processMessages(ctx context.Context, messageChan chan []byte) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("exiting client loop")
			return
		case msg := <-messageChan:
			log.Debugf("consumer received message %s", string(msg))
			processMessage(ctx, msg)
		}
	}
}

func processMessage(ctx context.Context, msg []byte) {
	if len(msg) != 4 {
		return
	}
	renicer.Renice(ctx, string(msg))
}
