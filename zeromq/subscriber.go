package zeromq

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
	receiveChan chan string
	out         chan<- string
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewSubscriber(topics []string, messageChan chan<- string) *Subscriber {
	return &Subscriber{
		topics:      topics,
		out:         messageChan,
		receiveChan: make(chan string, 256),
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
		msg, err := s.sock.Recv()
		if err != nil {
			if err.Error() == "context canceled" {
				return
			}
			log.Errorf("failed to receive message: %s", err)
			continue
		}
		log.Debugf("received [%s]", string(msg))
		s.receiveChan <- string(msg)
	}
}

func (s *Subscriber) receiveMessagesChan() {
	for {
		select {
		case <-s.ctx.Done():
			log.Debug("exiting message receiver")
			return
		case msg := <-s.receiveChan:
			s.out <- msg
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
