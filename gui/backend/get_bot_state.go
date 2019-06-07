package backend

import (
	"fmt"
	"net/http"

	"github.com/stellar/kelp/support/kelpos"
)

func (s *APIServer) getBotState(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in getBotState: %s\n", e))
		return
	}

	state, e := s.doGetBotState(botName)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error when getting bot state in getBotState: %s\n", e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s\n", state)))
}

func (s *APIServer) doGetBotState(botName string) (kelpos.BotState, error) {
	b, e := s.kos.GetBot(botName)
	if e != nil {
		return kelpos.InitState(), fmt.Errorf("error when getting bot state: %s", e)
	}
	return b.State, nil
}
