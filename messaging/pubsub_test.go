package messaging_test

import (
	"context"
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/messaging"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

var _ = Describe("Messaging", func() {
	Context("pubsub", func() {
		It("should send messages from publisher to subscriber", func() {
			const topic = "test"
			const msg = "foo"
			viper.Set(config.LogLevel, "trace")
			log.SetLevel(log.TraceLevel)

			publisher, err := messaging.NewPublisher()
			Expect(err).To(BeNil())
			defer publisher.Close()

			msgChan := make(chan messaging.PubSubMessage, 5)
			subscriber := messaging.NewSubscriber([]string{topic}, msgChan)
			defer subscriber.Close()
			subscriber.Start(context.Background())

			var msgCount int
			stopChan := make(chan bool)

			go func() {
				for {
					select {
					case <-msgChan:
						msgCount++
					case <-stopChan:
						return
					}
				}
			}()

			i := 5
			for i > 0 {
				message := messaging.NewPubSubMessage(topic, fmt.Sprintf("%s-%d", msg, i))
				err := publisher.Publish(message)
				Expect(err).ToNot(HaveOccurred())
				i--
				time.Sleep(10 * time.Millisecond)
			}

			Eventually(func() int {
				return msgCount
			}, "5s").Should(BeNumerically(">", 0))

			stopChan <- true
		})
	})
})
