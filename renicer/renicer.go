package renicer

import (
	"context"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	renicers = make(map[string]*renicer)
	revertDelay = time.Duration(viper.GetInt(config.NiceRevertDelayMinutes)) * time.Minute
)

type renicer struct {
	ctx context.Context
	cancel context.CancelFunc
	user string
	chain string
}

func init() {
	go awaitStop()
}

func Renice(chainID string) {
	chainID = strings.ToUpper(chainID)
	user := viper.GetString(chainID)

	rn, ok := renicers[chainID]
	if !ok {
		rn := &renicer{
			user: user,
			chain: chainID,
		}
		renicers[chainID] = rn
	}
	rn.renice()
}

func (rn *renicer) renice() {
	alreadyReniced := rn.ctx != nil
	ctx, cancel := context.WithTimeout(context.Background(), revertDelay)
	rn.ctx = ctx
	rn.cancel = cancel

	if alreadyReniced {
		// in effect, we just extended the timeout
		return
	}

	niceValue := viper.GetInt(config.NiceValue)
	if err := rn.runRenice(niceValue); err != nil {
		log.Errorf("failed to renice %s user for relay chain %s: %s", rn.user, rn.chain, err)
		rn.cancel()
		return
	}

	go rn.waitForTimeout()
}

func (rn *renicer) waitForTimeout() {
	for {
		select {
		case <- rn.ctx.Done():
			rn.cancel()
			go rn.revertNice()
			return
		}
	}
}

func (rn *renicer) revertNice() {
	if err := rn.runRenice(0); err != nil {
		log.Errorf("failed to revert niceness of user %s for chain %s: %s", rn.user, rn.chain, err)
	}
}

func (rn *renicer) runRenice(value int) error {
	strValue := strconv.Itoa(value)
	cmd := exec.Command("renice", "-n", strValue, "-u", rn.user)
	if err := cmd.Run(); err != nil {
		output, err2 := cmd.CombinedOutput()
		if err2 != nil {
			log.Errorf("failed to collect renice command output: %s", err2)
		} else {
			err = errors.Wrap(err, string(output))
		}
		return err
	}
	return nil
}

func awaitStop() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	for _, rn := range renicers {
		if rn.cancel != nil {
			rn.cancel()
			go rn.runRenice(0)
		}
	}
}
