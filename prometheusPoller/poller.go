package prometheusPoller

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/zeromq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	publisher   *zeromq.Publisher
	metricsUrl  string
	relayCounts = make(map[string]int)
	re          = regexp.MustCompile(`pocketcore_service_relay_count_for_([0-9A-F]{4}) (\d+)`) // ([0-9A-Z]{4}) (\d+)`)
)

func Start(ctx context.Context) error {
	var err error
	publisher, err = zeromq.NewPublisher()
	if err != nil {
		return err
	}
	defer publisher.Close()

	metricsUrl = fmt.Sprintf("http://127.0.0.1:%d/metrics", viper.GetInt(config.PrometheusPort))

	log.Infof("starting prometheus poller with metrics URL: %s", metricsUrl)

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping prometheus poller")
			return nil
		case <-ticker.C:
			poll()
		}
	}
}

func poll() {
	res, err := http.Get(metricsUrl)
	if err != nil {
		log.Errorf("failed to get metrics from prometheus server: %s", err)
		return
	}
	if res.StatusCode != 200 {
		log.Errorf("prometheus server returned status %d %s", res.StatusCode, res.Status)
		return
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("failed to read metrics response body: %s", err)
		return
	}

	messageChains := processPollData(data)
	publish(messageChains)
}

func processPollData(data []byte) []string {
	matches := re.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		return nil
	}

	var messageChains []string
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		chainID := m[1]
		relayCountStr := m[2]
		relayCount, err := strconv.Atoi(relayCountStr)
		if err != nil {
			return nil
		}

		if existingCount, ok := relayCounts[chainID]; ok {
			if existingCount != relayCount {
				relayCounts[chainID] = relayCount
				messageChains = append(messageChains, chainID)
			}
		} else {
			log.Infof("poller found chain %s", chainID)
			relayCounts[chainID] = relayCount
		}
	}
	return messageChains
}

func publish(messageChains []string) {
	var has0001 bool
	for _, chainID := range messageChains {
		if chainID == "0001" {
			has0001 = true
		}
		log.Infof("poller publishing message %s", chainID)
		if err := publisher.Publish(chainID, chainID); err != nil {
			log.Errorf("failed to publish %s: %s", chainID, err)
			return
		}
	}
	if len(messageChains) > 0 && !has0001 {
		// boost pocket too
		log.Debug("publishing message 0001")
		if err := publisher.Publish("0001", "0001"); err != nil {
			log.Errorf("failed to publish 0001: %s", err)
			return
		}
	}
}
