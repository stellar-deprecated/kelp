package backend

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stellar/kelp/support/kelpos"
)

func (s *APIServer) stopBot(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in stopBot: %s\n", e))
		return
	}

	e = s.doStopBot(botName)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error stopping bot: %s\n", e))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) doStopBot(botName string) error {
	e := s.kos.AdvanceBotState(botName, kelpos.BotStateRunning)
	if e != nil {
		return fmt.Errorf("error advancing bot state: %s\n", e)
	}

	e = s.kos.Stop(botName)
	if e != nil {
		return fmt.Errorf("error when killing bot %s: %s\n", botName, e)
	}
	log.Printf("stopped bot '%s'\n", botName)

	var numIterations uint8 = 1
	e = s.doStartBot(botName, "delete", &numIterations, func() {
		s.deleteFinishCallback(botName)
	})
	if e != nil {
		return fmt.Errorf("error when deleting bot orders %s: %s\n", botName, e)
	}
	return nil
}

func (s *APIServer) deleteFinishCallback(botName string) {
	log.Printf("deleted offers for bot '%s'\n", botName)

	e := s.kos.AdvanceBotState(botName, kelpos.BotStateStopping)
	if e != nil {
		log.Printf("error advancing bot state: %s\n", e)
	}
}
