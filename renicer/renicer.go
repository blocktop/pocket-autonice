package renicer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	renicers    = make(map[string]*renicer)
	initialized = false
)

type renicer struct {
	ctx         context.Context
	cancel      context.CancelFunc
	user        string
	chain       string
	dryRun      bool
	revertDelay time.Duration
}

func Renice(ctx context.Context, chainID string) {
	dryRun := viper.GetBool("dry_run")
	if !initialized {
		initialized = true
		go awaitStop(ctx, dryRun)
	}

	chainID = strings.ToUpper(chainID)
	user := getUserForChainID(chainID)
	if user == nil {
		log.Debugf("chain %s not configured; ignoring", chainID)
		return
	}

	rn, ok := renicers[chainID]
	if !ok {
		rn = &renicer{
			user:        *user,
			chain:       chainID,
			dryRun:      dryRun,
			revertDelay: time.Duration(viper.GetInt(config.NiceRevertDelayMinutes)) * time.Minute,
		}
		renicers[chainID] = rn
	}
	rn.renice(ctx)
}

func GetNiceValue(chainID string) (int, error) {
	chainID = strings.ToUpper(chainID)
	user := getUserForChainID(chainID)
	if user == nil {
		return 0, fmt.Errorf("chainID %s is not configured", chainID)
	}
	sudo := viper.GetBool(config.RunWithSudo)
	cmdName := "ps"
	args := []string{"-u", *user, "-o", "ni="}
	if sudo {
		cmdName = "sudo"
		args = append([]string{"ps"}, args...)
	}
	cmd := exec.Command(cmdName, args...)
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

func (rn *renicer) renice(ctx context.Context) {
	alreadyReniced := rn.ctx != nil
	ctx, cancel := context.WithTimeout(ctx, rn.revertDelay)
	rn.ctx = ctx
	rn.cancel = cancel

	if alreadyReniced {
		log.Debugf("reset revert timeer for chain %s", rn.chain)
		return
	}

	niceValue := viper.GetInt(config.NiceValue)
	renicing := "renicing"
	if rn.dryRun {
		renicing = "[DRY RUN] would renice"
	}
	log.Infof("%s chain %s (%s) to %d until %s of no activity", renicing, rn.chain, rn.user, niceValue,
		rn.revertDelay.String())
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
	reverting := "reverting"
	if rn.dryRun {
		reverting = "[DRY RUN] would revert"
	}
	log.Infof("%s niceness of chain %s (%s)) to 0", reverting, rn.chain, rn.user)
	if err := rn.runRenice(0); err != nil {
		log.Errorf("failed to revert niceness of user %s for chain %s: %s", rn.user, rn.chain, err)
	}
}

func (rn *renicer) runRenice(value int) error {
	if rn.dryRun {
		return nil
	}

	strValue := strconv.Itoa(value)
	sudo := viper.GetBool(config.RunWithSudo)
	cmdName := "renice"
	args := []string{"-n", strValue, "-u", rn.user}
	if sudo {
		cmdName = "sudo"
		args = append([]string{"renice"}, args...)
	}
	cmd := exec.Command(cmdName, args...)
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

func awaitStop(ctx context.Context, dryRun bool) {
	<-ctx.Done()

	reverting := "reverting"
	if dryRun {
		reverting = "[DRY RUN] would revert"
	}
	log.Infof("%s nice to 0 on all chains", reverting)
	for _, rn := range renicers {
		if rn.cancel != nil {
			rn.cancel()
			go rn.runRenice(0)
		}
	}
}

func getUserForChainID(chainID string) *string {
	chains := viper.GetStringMapString(config.Chains)
	if user, ok := chains[chainID]; ok {
		return &user
	}
	return nil
}
