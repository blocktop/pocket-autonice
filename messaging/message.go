package messaging

import (
	"fmt"
	"strings"
)

type PubSubMessage struct {
	topic   string
	message string
}

func NewPubSubMessage(topic, message string) PubSubMessage {
	return PubSubMessage{
		topic:   topic,
		message: message,
	}
}

func (m PubSubMessage) Topic() string {
	return m.topic
}

func (m PubSubMessage) Message() string {
	return m.message
}

func (m PubSubMessage) String() string {
	return fmt.Sprintf("%s %s", m.topic, m.message)
}

func (m PubSubMessage) Marshal() []byte {
	return []byte(m.String())
}

func UnmarshalPubSubMessage(data []byte) PubSubMessage {
	s := string(data)
	i := strings.Index(s, " ")
	m := PubSubMessage{}
	if i == -1 {
		m.topic = s
		m.message = s
	} else {
		m.topic = s[:i]
		m.message = s[i+1:]
	}
	return m
}
