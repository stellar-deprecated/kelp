package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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

	strategy := "buysell"
	filenamePair := model.GetBotFilenames(botName, strategy)
	logPrefix := model.GetLogPrefix(botName, strategy)
	command := fmt.Sprintf("trade -c %s/%s -s %s -f %s/%s -l %s/%s", s.configsDir, filenamePair.Trader, strategy, s.configsDir, filenamePair.Strategy, s.logsDir, logPrefix)
	log.Printf("run command for bot '%s': %s\n", botName, command)

	go func(kelpCmdString string, name string) {
		output, e := s.runKelpCommand(kelpCmdString)
		if e != nil {
			fmt.Printf("error when starting bot '%s': %s", name, e)
		}
		fmt.Printf("finished start bot command with result: %s\n", string(output))
	}(command, botName)

	w.WriteHeader(http.StatusOK)
}
