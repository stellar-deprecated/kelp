package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/stellar/kelp/gui/model"
)

func (s *APIServer) startBot(w http.ResponseWriter, r *http.Request) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when reading request input: %s\n", e)))
		return
	}

	botName := string(botNameBytes)
	e = s.doStartBot(botName, "buysell", nil)
	if e != nil {
		log.Printf("error starting bot: %s", e)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) doStartBot(botName string, strategy string, iterations *uint8) error {
	filenamePair := model.GetBotFilenames(botName, strategy)
	logPrefix := model.GetLogPrefix(botName, strategy)
	command := fmt.Sprintf("trade -c %s/%s -s %s -f %s/%s -l %s/%s", s.configsDir, filenamePair.Trader, strategy, s.configsDir, filenamePair.Strategy, s.logsDir, logPrefix)
	if iterations != nil {
		command = fmt.Sprintf("%s --iter %d", command, *iterations)
	}
	log.Printf("run command for bot '%s': %s\n", botName, command)

	c, e := s.runKelpCommandBackground(botName, command)
	if e != nil {
		return fmt.Errorf("could not start bot %s: %s", botName, e)
	}

	go func(kelpCommand *exec.Cmd, name string) {
		defer s.kos.SafeUnregister(name)

		if kelpCommand != nil {
			e := kelpCommand.Wait()
			if e != nil {
				if strings.Contains(e.Error(), "signal: terminated") {
					log.Printf("terminated start bot command for bot '%s' with strategy '%s'\n", name, strategy)
					return
				}
				log.Printf("error when starting bot '%s' with strategy '%s': %s\n", name, strategy, e)
				return
			}
		}
		log.Printf("finished start bot command for bot '%s' with strategy '%s'\n", name, strategy)
	}(c, botName)

	return nil
}
