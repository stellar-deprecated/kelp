package backend

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
		s.writeKelpError(w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelWarning,
			fmt.Sprintf("unable to stop bot: %s\n", e),
		))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) doStopBot(botName string) error {
	e := s.kos.AdvanceBotState(botName, kelpos.BotStateRunning)
	if e != nil {
		return fmt.Errorf("error advancing bot state: %s", e)
	}

	e = s.kos.Stop(botName)
	if e != nil {
		return fmt.Errorf("error when killing bot %s: %s", botName, e)
	}
	log.Printf("stopped bot '%s'\n", botName)

	var numIterations uint8 = 1
	e = s.doStartBot(botName, "delete", &numIterations, func() {
		eInner := s.deleteFinishCallback(botName)
		if eInner != nil {
			s.addKelpErrorToMap(makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelWarning,
				fmt.Sprintf("error running deleteFinishCallback when stopping bot: %s", eInner),
			).KelpError)
			log.Printf("error running deleteFinishCallback when stopping bot: %s", eInner)
		}
	})
	if e != nil {
		return fmt.Errorf("error when deleting bot orders %s: %s", botName, e)
	}
	return nil
}

func (s *APIServer) deleteFinishCallback(botName string) error {
	log.Printf("deleted offers for bot '%s'\n", botName)

	e := s.kos.AdvanceBotState(botName, kelpos.BotStateStopping)
	if e != nil {
		return fmt.Errorf("error advancing bot state when manually attempting to stop bot: %s", e)
	}
	return nil
}
