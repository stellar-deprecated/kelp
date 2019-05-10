package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	s.doStartBot(botName, "buysell", nil)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) doStartBot(botName string, strategy string, iterations *uint8) {
	filenamePair := model.GetBotFilenames(botName, strategy)
	logPrefix := model.GetLogPrefix(botName, strategy)
	command := fmt.Sprintf("trade -c %s/%s -s %s -f %s/%s -l %s/%s", s.configsDir, filenamePair.Trader, strategy, s.configsDir, filenamePair.Strategy, s.logsDir, logPrefix)
	if iterations != nil {
		command = fmt.Sprintf("%s --iter %d", command, *iterations)
	}
	log.Printf("run command for bot '%s': %s\n", botName, command)

	go func(kelpCmdString string, name string) {
		// runKelpCommand is blocking
		_, e := s.runKelpCommand(kelpCmdString)
		if e != nil {
			if strings.Contains(e.Error(), "signal: terminated") {
				fmt.Printf("terminated start bot command for bot '%s' with strategy '%s'\n", name, strategy)
				return
			}
			fmt.Printf("error when starting bot '%s' with strategy '%s': %s\n", name, strategy, e)
			return
		}
		fmt.Printf("finished start bot command for bot '%s' with strategy '%s'\n", name, strategy)
	}(command, botName)
}
