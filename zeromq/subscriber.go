package zeromq

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zeromq/goczmq"
	"time"
)

type Subscriber struct {
	channeler *goczmq.Channeler
	topic     string
	out       chan<- []byte
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewSubscriber(topic string, messageChan chan<- []byte) *Subscriber {
	return &Subscriber{
		topic: topic,
		out:   messageChan,
	}
}

func (s *Subscriber) Start() {
	if s.channeler != nil {
		return
	}

	endpoint := getTCPEndpoint()
	ch := goczmq.NewSubChanneler(endpoint, s.topic)
	s.channeler = ch

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go s.receiveMessages()
}

func (s *Subscriber) receiveMessages() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-s.ctx.Done():
			log.Debug("exiting message receiver")
			return
		case <-ticker.C:
			// ensure subscription in case publisher stops and restarts
			s.channeler.Subscribe(s.topic)
		case data := <-s.channeler.RecvChan:
			log.Debugf("received %#v", data)
			if len(data) > 1 {
				msg := data[len(data)-1]
				s.out <- msg
			}
		}
	}
}

func (s *Subscriber) Close() {
	s.cancel()
	if s.channeler != nil {
		s.channeler.Destroy()
		s.channeler = nil
	}
}
