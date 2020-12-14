package backend

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stellar/kelp/support/kelpos"
)

func (s *APIServer) getBotState(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		// we do not have the botName so we cannot throw a bot specific error here
		s.writeError(w, fmt.Sprintf("error in getBotState: %s\n", e))
		return
	}

	state, e := s.doGetBotState(botName)
	if e != nil {
		s.writeKelpError(w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelWarning,
			fmt.Sprintf("unable to get bot state: %s\n", e),
		))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s\n", state)))
}

func (s *APIServer) doGetBotState(botName string) (kelpos.BotState, error) {
	b, e := s.kos.GetBot(botName)
	if e != nil {
		return kelpos.InitState(), fmt.Errorf("unable to get bot state: %s", e)
	}
	log.Printf("bots available: %v", s.kos.RegisteredBots())
	return b.State, nil
}
