package zeromq

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/blocktop/pocket-autonice/config"
	zmq "github.com/go-zeromq/zmq4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Publisher struct {
	sock zmq.Socket
}

func NewPublisher() (*Publisher, error) {
	p := &Publisher{}
	if err := p.createSock(); err != nil {
		return nil, errors.Wrap(err, "fatal error occurred creating publisher socket")
	}
	return p, nil
}

func (p *Publisher) Publish(msg, topic string) error {
	if p.sock == nil {
		return fmt.Errorf("publisher socket has been closed")
	}
	m := zmq.NewMsgFrom([]byte(topic), []byte(msg))
	log.Debugf("publisher sending [%s]", m.String())
	err := p.sock.Send(m)
	if err != nil {
		err = errors.Wrap(err, "error occurred publishing message")
		log.Errorf(err.Error())
		return err
	}
	return nil
}

func (p *Publisher) createSock() error {
	sock := zmq.NewPub(context.Background())
	endpoint := fmt.Sprintf("tcp://%s", viper.GetString(config.PublisherBindAddress))
	if err := sock.Listen(endpoint); err != nil {
		return errors.Wrap(err, "failed to bind zmq publisher socket")
	}
	p.sock = sock

	// give publishers time to see the subscriptions
	time.Sleep(time.Second)

	return nil
}

func (p *Publisher) Close() {
	if p.sock != nil {
		p.sock.Close()
		p.sock = nil
	}
}

func makePubMessage(msg []byte, topic string) [][]byte {
	return [][]byte{
		[]byte(topic),
		msg,
	}
}
