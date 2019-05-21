package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *APIServer) stopBot(w http.ResponseWriter, r *http.Request) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when reading request input: %s\n", e)))
		return
	}

	botName := string(botNameBytes)
	e = s.stopCommand(botName)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when killing bot %s: %s\n", botName, e)))
		return
	}
	log.Printf("stopped bot '%s'\n", botName)

	var numIterations uint8 = 1
	e = s.doStartBot(botName, "delete", &numIterations)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when deleting bot ortders %s: %s\n", botName, e)))
		return
	}

	w.WriteHeader(http.StatusOK)
}
