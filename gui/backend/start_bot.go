package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/support/kelpos"
)

func (s *APIServer) startBot(w http.ResponseWriter, r *http.Request) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}

	botName := string(botNameBytes)
	e = s.doStartBot(botName, "buysell", nil, nil)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error starting bot: %s\n", e))
		return
	}

	e = s.kos.AdvanceBotState(botName, kelpos.BotStateStopped)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error advancing bot state: %s\n", e))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) doStartBot(botName string, strategy string, iterations *uint8, maybeFinishCallback func()) error {
	filenamePair := model2.GetBotFilenames(botName, strategy)
	logPrefix := model2.GetLogPrefix(botName, strategy)

	// use relative paths here so it works under windows. In windows we use unix paths to reference the config files since it is
	// started under the linux subsystem, but it is a windows binary so uses the windows naming scheme (C:\ etc.). Therefore we need
	// to either find a regex replacement to convert from unix to windows (/mnt/c -> C:\) or we can use relative paths which we did.
	// Note that /mnt/c is unlikely to be valid in windows (but is valid in the linux subsystem) since it's usually prefixed by the
	// volume (C:\ etc.), which is why relative paths works so well here as it avoids this confusion.
	traderRelativeConfigPath, e := s.configsDir.Join(filenamePair.Trader).RelFromPath(s.basepath)
	if e != nil {
		return fmt.Errorf("unable to get relative path of trader config file from basepath: %s", e)
	}

	stratRelativeConfigPath, e := s.configsDir.Join(filenamePair.Strategy).RelFromPath(s.basepath)
	if e != nil {
		return fmt.Errorf("unable to get relative path of strategy config file from basepath: %s", e)
	}

	logRelativePrefix, e := s.logsDir.Join(logPrefix).RelFromPath(s.basepath)
	if e != nil {
		return fmt.Errorf("unable to get relative log prefix from basepath: %s", e)
	}

	command := fmt.Sprintf("trade -c %s -s %s -f %s -l %s --ui",
		traderRelativeConfigPath.Unix(),
		strategy,
		stratRelativeConfigPath.Unix(),
		logRelativePrefix.Unix(),
	)
	if iterations != nil {
		command = fmt.Sprintf("%s --iter %d", command, *iterations)
	}
	if s.noHeaders {
		command = fmt.Sprintf("%s --no-headers", command)
	}
	if s.ccxtRestUrl != "" {
		command = fmt.Sprintf("%s --ccxt-rest-url %s", command, s.ccxtRestUrl)
	}
	log.Printf("run command for bot '%s': %s\n", botName, command)

	p, e := s.runKelpCommandBackground(botName, command)
	if e != nil {
		return fmt.Errorf("could not start bot %s: %s", botName, e)
	}

	go func(kelpCommand *exec.Cmd, name string) {
		defer s.kos.SafeUnregister(name)

		if kelpCommand == nil {
			log.Printf("kelpCommand was nil for bot '%s' with strategy '%s'\n", name, strategy)
			return
		}

		e := kelpCommand.Wait()
		if e != nil {
			if strings.Contains(e.Error(), "signal: terminated") {
				log.Printf("terminated start bot command for bot '%s' with strategy '%s'\n", name, strategy)
				return
			}
			log.Printf("error when starting bot '%s' with strategy '%s': %s\n", name, strategy, e)
			return
		}

		log.Printf("finished start bot command for bot '%s' with strategy '%s'\n", name, strategy)
		if maybeFinishCallback != nil {
			maybeFinishCallback()
		}
	}(p.Cmd, botName)

	return nil
}
