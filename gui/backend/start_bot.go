package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/stellar/kelp/gui/model"
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
	filenamePair := model.GetBotFilenames(botName, strategy)
	logPrefix := model.GetLogPrefix(botName, strategy)
	command := fmt.Sprintf("trade -c %s/%s -s %s -f %s/%s -l %s/%s", s.configsDir, filenamePair.Trader, strategy, s.configsDir, filenamePair.Strategy, s.logsDir, logPrefix)
	if iterations != nil {
		command = fmt.Sprintf("%s --iter %d", command, *iterations)
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
