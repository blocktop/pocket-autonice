package prometheusPoller

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Poller", func() {
	Context("processPollData", func() {
		It("should initialize and then message changes in relays", func() {
			messageChains := processPollData(data1)
			Expect(messageChains).To(HaveLen(0))

			messageChains = processPollData(data2)
			Expect(messageChains).To(HaveLen(2))
			Expect(messageChains).To(ContainElement("0009"))
			Expect(messageChains).To(ContainElement("0040"))
		})
	})
})
