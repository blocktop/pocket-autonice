package zeromq

import (
	log "github.com/sirupsen/logrus"
	"github.com/zeromq/goczmq"
)

type Subscriber struct {
	channeler *goczmq.Channeler
	stopChan  chan bool
	topic     string
	out       chan<- []byte
}

func NewSubscriber(topic string, messageChan chan<- []byte) *Subscriber {
	return &Subscriber{
		topic:    topic,
		stopChan: make(chan bool, 1),
		out:      messageChan,
	}
}

func (s *Subscriber) Start() {
	if s.channeler != nil {
		return
	}

	endpoint := getTCPEndpoint()
	ch := goczmq.NewSubChanneler(endpoint, s.topic)
	s.channeler = ch

	go s.receiveMessages()
}

func (s *Subscriber) receiveMessages() {
	for {
		select {
		case <-s.stopChan:
			log.Debug("exiting message receiver")
			return
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
	s.stopChan <- true
	if s.channeler != nil {
		s.channeler.Destroy()
		s.channeler = nil
	}
}
