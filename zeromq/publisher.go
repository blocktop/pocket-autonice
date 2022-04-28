package zeromq

import (
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zeromq/goczmq"
	"time"
)

type Publisher struct {
	sock *goczmq.Sock
}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) Publish(msg []byte, topic string) error {
	if p.sock == nil {
		if err := p.createSock(); err != nil {
			log.Fatalf("fatal error occurred creating publisher socket: %s", err)
		}
	}
	data := makePubMessage(msg, topic)
	if err := p.sock.SendMessage(data); err != nil {
		log.Errorf("error occurred publishing message: %s", err)
		return err
	}
	return nil
}

func (p *Publisher) createSock() error {
	endpoint := getTCPEndpoint()
	sock, err := goczmq.NewPub(endpoint)
	if err != nil {
		return err
	}
	sock.SetLinger(0)
	p.sock = sock

	// give publishers time to see the subscription
	time.Sleep(time.Second)

	return nil
}

func (p *Publisher) Close() {
	if p.sock != nil {
		p.sock.Destroy()
		p.sock = nil
	}
}

func getTCPEndpoint() string {
	return fmt.Sprintf("tcp://%s", viper.GetString(config.ZeroMQAddress))
}

func makePubMessage(msg []byte, topic string) [][]byte {
	return [][]byte{
		[]byte(topic),
		msg,
	}
}
