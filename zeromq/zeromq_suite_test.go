package zeromq_test

import (
	"github.com/blocktop/pocket-autonice/config"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZeromq(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zeromq Suite")
}

var _ = BeforeEach(func() {
	config.InitConfig()
})
