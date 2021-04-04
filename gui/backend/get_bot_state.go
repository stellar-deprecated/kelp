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
	// TODO fetch userData and pass in here
	userData := UserData{}

	state, e := s.doGetBotState(userData, botName)
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
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

func (s *APIServer) doGetBotState(userData UserData, botName string) (kelpos.BotState, error) {
	ubd := s.kos.BotDataForUser(userData.toUser())
	b, e := ubd.GetBot(botName)
	if e != nil {
		return kelpos.InitState(), fmt.Errorf("unable to get bot state: %s", e)
	}
	log.Printf("bots available for user (%s): %v", userData.String(), ubd.RegisteredBots())
	return b.State, nil
}
