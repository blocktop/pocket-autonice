package renicer

import (
	"bytes"
	"context"
	"fmt"
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
	renicers    = make(map[string]*renicer)
	revertDelay = time.Duration(viper.GetInt(config.NiceRevertDelayMinutes)) * time.Minute
)

type renicer struct {
	ctx    context.Context
	cancel context.CancelFunc
	user   string
	chain  string
}

func init() {
	go awaitStop()
}

func Renice(chainID string) {
	chainID = strings.ToUpper(chainID)
	user := getUserForChainID(chainID)
	if user == nil {
		return
	}

	rn, ok := renicers[chainID]
	if !ok {
		rn := &renicer{
			user:  *user,
			chain: chainID,
		}
		renicers[chainID] = rn
	}
	rn.renice()
}

func GetNiceValue(chainID string) (int, error) {
	chainID = strings.ToUpper(chainID)
	user := getUserForChainID(chainID)
	if user == nil {
		return 0, fmt.Errorf("chainID %s is not configured", chainID)
	}
	cmd := exec.Command("ps", "-u", *user, "-o", "ni=")
	var outData bytes.Buffer
	cmd.Stdout = &outData

	if err := cmd.Run(); err != nil {
		return 0, errors.Wrap(err, "ps command failed")
	}

	outStr := outData.String()
	outs := strings.Split(outStr, "\n")
	if len(outs) == 0 {
		return 0, fmt.Errorf("no nice value was found")
	}
	out := strings.TrimSpace(outs[0])

	nice, err := strconv.Atoi(string(out))
	if err != nil {
		return 0, fmt.Errorf("unable to convert output '%s' to an integer: %s", string(out), err)
	}

	return nice, nil
}

func (rn *renicer) renice() {
	alreadyReniced := rn.ctx != nil
	ctx, cancel := context.WithTimeout(context.Background(), revertDelay)
	rn.ctx = ctx
	rn.cancel = cancel

	if alreadyReniced {
		log.Debugf("reset revert timeer for chain %s", rn.chain)
		return
	}

	niceValue := viper.GetInt(config.NiceValue)
	log.Infof("renicing chain %s (%s) to %d", rn.chain, rn.user, niceValue)
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
		case <-rn.ctx.Done():
			rn.cancel()
			go rn.revertNice()
			return
		}
	}
}

func (rn *renicer) revertNice() {
	log.Infof("reverting niceness of chain %s (%s)) to 0", rn.chain, rn.user)
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

	log.Info("reverting nice to 0 on all chains")
	for _, rn := range renicers {
		if rn.cancel != nil {
			rn.cancel()
			go rn.runRenice(0)
		}
	}
}

func getUserForChainID(chainID string) *string {
	if !viper.IsSet(chainID) {
		return nil
	}

	user := viper.GetString(chainID)
	return &user
}
