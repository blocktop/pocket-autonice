package zeromq

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/blocktop/pocket-autonice/config"
	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Publisher struct {
	sock *zmq.Socket
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

	_, err := p.sock.Send(topic, zmq.SNDMORE)
	if err == nil {
		_, err = p.sock.Send(msg, 0)
	}
	if err != nil {
		err = errors.Wrap(err, "error occurred publishing message")
		log.Errorf(err.Error())
		return err
	}
	return nil
}

func (p *Publisher) createSock() error {
	sock, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return errors.Wrap(err, "failed to create zmq publisher socket")
	}
	if err = sock.SetLinger(0); err != nil {
		return errors.Wrap(err, "failed to set linger on zmq publisher socket")
	}
	if strings.ToLower(viper.GetString(config.LogLevel)) == "trace" {
		const monitorAddr = "inproc://monitor.pub"
		if err = sock.Monitor(monitorAddr, zmq.EVENT_ALL); err != nil {
			return errors.Wrap(err, "failed to configure monitor on zmq publisher socket")
		}
		go monitorSocket(monitorAddr, "PUB")
		time.Sleep(time.Second)
	}
	endpoint := fmt.Sprintf("tcp://%s", viper.GetString(config.PublisherBindAddress))
	if err = sock.Bind(endpoint); err != nil {
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
