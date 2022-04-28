package client

import (
	"github.com/blocktop/pocket-autonice/renicer"
	"github.com/blocktop/pocket-autonice/zeromq"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var (
	subscriber *zeromq.Subscriber
)

func Start() {
	pubsubTopic := viper.GetString(PubSubTopic)
	messageChan := make(chan []byte, 256)
	subscriber = zeromq.NewSubscriber(pubsubTopic, messageChan)
	defer subscriber.Close()

	stopChan := make(chan bool)

	go processMessages(messageChan, stopChan)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	stopChan <- true
}

func processMessages(messageChan chan []byte, stopChan chan bool) {
	for {
		select {
		case msg := <-messageChan:
			processMessage(msg)
		case <-stopChan:
			return
		}
	}
}

func processMessage(msg []byte) {
	if len(msg) != 4 {
		return
	}
	renicer.Renice(string(msg))
}
