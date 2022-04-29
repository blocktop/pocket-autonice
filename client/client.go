package client

import (
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/renicer"
	"github.com/blocktop/pocket-autonice/zeromq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var (
	subscriber *zeromq.Subscriber
)

func Start() {
	pubsubTopic := viper.GetString(config.PubSubTopic)
	messageChan := make(chan []byte, 256)
	subscriber = zeromq.NewSubscriber(pubsubTopic, messageChan)
	defer subscriber.Close()

	stopChan := make(chan bool)

	log.Info("starting message consumer")

	go processMessages(messageChan, stopChan)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("stopping message consumer")
	stopChan <- true
}

func processMessages(messageChan chan []byte, stopChan chan bool) {
	for {
		select {
		case <-stopChan:
			log.Debug("exiting client loop")
			return
		case msg := <-messageChan:
			log.Debugf("consumer received message %s", string(msg))
			processMessage(msg)
		}
	}
}

func processMessage(msg []byte) {
	if len(msg) != 4 {
		return
	}
	renicer.Renice(string(msg))
}
