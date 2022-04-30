package prometheusPoller

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPrometheusPoller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PrometheusPoller Suite")
}

var (
	data1 = []byte(`
pocketcore_service_relay_count_for_0009 19111
pocketcore_service_relay_count_for_0021 2222
pocketcore_service_relay_count_for_0040 20888
`)
	data2 = []byte(`
pocketcore_service_relay_count_for_0009 19222
pocketcore_service_relay_count_for_0021 2222
pocketcore_service_relay_count_for_0040 20999
`)
)
