package zeromq

import (
	"context"
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	zmq "github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

type Subscriber struct {
	sock        *zmq.Socket
	topic       string
	receiveChan chan string
	out         chan<- string
	ctx         context.Context
	cancel      context.CancelFunc
	canceled    bool
}

func NewSubscriber(topic string, messageChan chan<- string) *Subscriber {
	return &Subscriber{
		topic:       topic,
		out:         messageChan,
		receiveChan: make(chan string, 256),
	}
}

func (s *Subscriber) Start() error {
	if s.sock != nil {
		return nil
	}

	zctx, err := zmq.NewContext()
	if err != nil {
		return errors.Wrap(err, "failed to create zmq context")
	}
	sock, err := zctx.NewSocket(zmq.SUB)
	if err != nil {
		return errors.Wrap(err, "failed to create zmq subscriber socket")
	}
	if err = sock.SetHeartbeatIvl(time.Second); err != nil {
		return errors.Wrap(err, "failed to set heartbeat interval (requires zmq >= 4.2")
	}
	if err = sock.SetReconnectIvl(time.Minute); err != nil {
		return errors.Wrap(err, "failed to set reconnect interval on zmq subscriber socket")
	}
	if err = sock.Connect(fmt.Sprintf("tcp://%s", viper.GetString(config.SubscriberAddress))); err != nil {
		return errors.Wrap(err, "failed to connect zmq subscriber socket")
	}
	if err = sock.SetSubscribe(s.topic); err != nil {
		return errors.Wrap(err, "failed to  set topic subscription")
	}

	s.sock = sock

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go s.receiveMessagesChan()
	go s.receiveMessages()

	return nil
}

func (s *Subscriber) receiveMessages() {
	for !s.canceled {
		address, err := s.sock.Recv(0)
		if err != nil {
			log.Errorf("failed to receive address: %s", err)
			continue
		}
		msg, err := s.sock.Recv(0)
		if err != nil {
			log.Errorf("failed to receive message: %s", err)
			continue
		}
		log.Debugf("received %s from %s", msg, address)
		s.receiveChan <- msg
	}
}

func (s *Subscriber) receiveMessagesChan() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-s.ctx.Done():
			log.Debug("exiting message receiver")
			s.canceled = true
			return
		case msg := <-s.receiveChan:
			s.out <- msg
		case <-ticker.C:
			s.sock.SetSubscribe("ping")
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
