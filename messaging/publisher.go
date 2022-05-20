package messaging

import (
	"fmt"
	"time"

	"github.com/blocktop/pocket-autonice/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	_ "go.nanomsg.org/mangos/v3/transport/tcp"
)

type Publisher struct {
	sock mangos.Socket
}

func NewPublisher() (*Publisher, error) {
	p := &Publisher{}
	if err := p.createSock(); err != nil {
		return nil, errors.Wrap(err, "fatal error occurred creating publisher socket")
	}
	return p, nil
}

func (p *Publisher) Publish(message PubSubMessage) error {
	if p.sock == nil {
		return fmt.Errorf("publisher socket has been closed")
	}
	log.Debugf("publisher sending [%s]", message)
	err := p.sock.Send(message.Marshal())
	if err != nil {
		err = errors.Wrap(err, "error occurred publishing message")
		log.Errorf(err.Error())
		return err
	}
	return nil
}

func (p *Publisher) createSock() error {
	sock, err := pub.NewSocket()
	if err != nil {
		return errors.Wrap(err, "failed to create publisher socket")
	}
	endpoint := fmt.Sprintf("tcp://%s", viper.GetString(config.PublisherBindAddress))
	if err := sock.Listen(endpoint); err != nil {
		return errors.Wrap(err, "failed to bind publisher socket")
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
