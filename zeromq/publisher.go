package zeromq

import (
	"fmt"
	"github.com/pkg/errors"
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
	zctx, err := zmq.NewContext()
	if err != nil {
		return errors.Wrap(err, "failed to create zmq context")
	}
	sock, err := zctx.NewSocket(zmq.PUB)
	if err != nil {
		return errors.Wrap(err, "failed to create zmq publisher socket")
	}
	if err = sock.SetLinger(100 * time.Millisecond); err != nil {
		return errors.Wrap(err, "failed to set linger on zmq publisher socket")
	}
	if err = sock.SetHeartbeatIvl(time.Second); err != nil {
		return errors.Wrap(err, "failed to set heartbeat interval (requires zmq >= 4.2)")
	}
	if err = sock.SetReconnectIvl(time.Minute); err != nil {
		return errors.Wrap(err, "failed to set reconnect interval")
	}
	endpoint := fmt.Sprintf("tcp://%s", viper.GetString(config.PublisherAddress))
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
