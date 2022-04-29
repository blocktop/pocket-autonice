package server

import (
	"encoding/json"
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/zeromq"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	svr         *http.Server
	publisher   *zeromq.Publisher
	pubsubTopic string
	shunting    bool
)

type MirrorData struct {
	RelayNetworkID string `json:"relay_network_id"`
}

func Start() {
	publisher = zeromq.NewPublisher()
	defer publisher.Close()

	pubsubTopic = viper.GetString(config.PubSubTopic)

	r := chi.NewRouter()
	r.Post("/", postHandler)

	addr := fmt.Sprintf("127.0.0.1:%d", viper.Get(config.ServerPort))

	svr = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	shunting = false

	log.Infof("starting server on %s", addr)
	go func() {
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %s", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("stopping server")

	if err := svr.Close(); err != nil {
		log.Errorf("server failed to close: %s", err)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if shunting {
		w.Write([]byte("server shunted"))
		return
	}
	uri := r.Header.Get("X-Original-URI")
	if uri == "" {
		shunting = true
		log.Errorf("mirror request must have X-Original-URI header; see documentation")
		w.Write([]byte("missing X-Original-URI header"))
		return
	}
	if uri != "/v1/client/relay" && uri != "/v1/client/sim" {
		w.Write([]byte("not relay or sim"))
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		w.Write([]byte("not json"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read %s mirror request body: %s", uri, err)
		w.Write([]byte("bad request body"))
		return
	}

	var data MirrorData
	if err = json.Unmarshal(body, &data); err != nil {
		log.Errorf("failed to unmarshal %s mirror body: %s", uri, err)
		w.Write([]byte("bad json"))
		return
	}

	if data.RelayNetworkID == "" || len(data.RelayNetworkID) != 4 {
		log.Errorf("bad relay network ID in %s request: %s", uri, data.RelayNetworkID)
		w.Write([]byte("bad relay network ID"))
		return
	}

	log.Debugf("publishing message %s", data.RelayNetworkID)
	if err = publisher.Publish([]byte(data.RelayNetworkID), pubsubTopic); err != nil {
		log.Errorf("failed to publish %s from %s: %s", data.RelayNetworkID, uri, err)
		w.Write([]byte("failed to publish"))
		return
	}
	// boost pocket too
	if err = publisher.Publish([]byte("0001"), pubsubTopic); err != nil {
		log.Errorf("failed to publish 0001 from %s: %s", uri, err)
		w.Write([]byte("failed to publish"))
		return
	}

	w.Write([]byte("OK"))
}
