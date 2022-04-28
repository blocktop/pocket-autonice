package zeromq_test

import (
	"fmt"
	"github.com/blocktop/pocket-autonice/zeromq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("∅MQ", func() {
	Context("pubsub", func() {
		It("should send messages from publisher to subscriber", func() {
			const topic = "test"
			const msg = "foo"

			publisher := zeromq.NewPublisher()
			defer publisher.Close()

			msgChan := make(chan []byte, 5)
			subscriber := zeromq.NewSubscriber(topic, msgChan)
			defer subscriber.Close()
			subscriber.Start()

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
				err := publisher.Publish([]byte(fmt.Sprintf("%s-%d", msg, i)), topic)
				Expect(err).ToNot(HaveOccurred())
				i--
				time.Sleep(10 * time.Millisecond)
			}

			Eventually(func() int {
				return msgCount
			}, "2s").Should(Equal(5))

			stopChan <- true
		})
	})
})
