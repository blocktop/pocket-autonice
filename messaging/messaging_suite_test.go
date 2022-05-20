package messaging_test

import (
	"github.com/blocktop/pocket-autonice/config"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMessaging(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Messaging Suite")
}

var _ = BeforeEach(func() {
	config.InitConfig()
})
