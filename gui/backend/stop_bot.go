package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/kelp/gui/model"
)

func (s *APIServer) stopBot(w http.ResponseWriter, r *http.Request) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when reading request input: %s\n", e)))
		return
	}
	botName := string(botNameBytes)
	strategy := "buysell"

	filenamePair := model.GetBotFilenames(botName, strategy)
	command := fmt.Sprintf("ps aux | grep %s | col | cut -d' ' -f2 | xargs kill", filenamePair.Trader)
	log.Printf("stop command for bot '%s': %s\n", botName, command)

	go func(cmdString string, name string) {
		// runKelpCommand is blocking
		_, e := runBashCommand(cmdString)
		if e != nil {
			if strings.Contains(e.Error(), "signal: terminated") {
				fmt.Printf("stopped bot '%s'\n", name)

				var numIterations uint8 = 1
				s.doStartBot(botName, "delete", &numIterations)
				return
			}
			fmt.Printf("error when stopping bot '%s': %s\n", name, e)
			return
		}
	}(command, botName)

	w.WriteHeader(http.StatusOK)
}
