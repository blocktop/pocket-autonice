package messaging

import (
	"context"
	"fmt"

	"github.com/blocktop/pocket-autonice/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	_ "go.nanomsg.org/mangos/v3/transport/tcp"
)

type Subscriber struct {
	sock        mangos.Socket
	topics      []string
	receiveChan chan []byte
	out         chan<- PubSubMessage
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewSubscriber(topics []string, messageChan chan<- PubSubMessage) *Subscriber {
	return &Subscriber{
		topics:      topics,
		out:         messageChan,
		receiveChan: make(chan []byte, 256),
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	if s.sock != nil {
		return nil
	}

	sock, err := sub.NewSocket()
	if err != nil {
		return errors.Wrap(err, "failed to create subscriber socket")
	}
	bindAddr := viper.GetString(config.SubscriberBindAddress)
	if bindAddr != "" {
		if err = sock.SetOption(mangos.OptionLocalAddr, bindAddr); err != nil {
			return errors.Wrap(err, "failed to set subscriber bind address")
		}
	}
	var endpoint string
	subPubAddr := viper.GetString(config.SubscriberPublisherAddress)
	endpoint = fmt.Sprintf("tcp://%s", subPubAddr)
	if err := sock.Dial(endpoint); err != nil {
		return errors.Wrapf(err, "failed to connect subscriber socket %s", endpoint)
	}
	for _, t := range s.topics {
		if err = sock.SetOption(mangos.OptionSubscribe, t); err != nil {
			return errors.Wrap(err, "failed to set topic subscription")
		}
	}

	s.sock = sock

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go s.receiveMessagesChan()
	go s.receiveMessages()

	return nil
}

func (s *Subscriber) receiveMessages() {
	for {
		recv, err := s.sock.Recv()
		if err != nil {
			if err.Error() == "context canceled" || err.Error() == "object closed" {
				return
			}
			log.Errorf("failed to receive message: %s", err)
			continue
		}
		s.receiveChan <- recv
	}
}

func (s *Subscriber) receiveMessagesChan() {
	for {
		select {
		case <-s.ctx.Done():
			log.Debug("exiting message receiver")
			return
		case data := <-s.receiveChan:
			message := UnmarshalPubSubMessage(data)
			log.Debugf("received [%s]", message)
			s.out <- message
		}
	}
}

func (s *Subscriber) Close() {
	s.cancel()
	if s.sock != nil {
		s.sock.Close()
		s.sock = nil
	}
}
