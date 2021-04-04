package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stellar/kelp/support/kelpos"
)

type getBotStateRequest struct {
	UserData UserData `json:"user_data"`
	BotName  string   `json:"bot_name"`
}

func (s *APIServer) getBotState(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		// we do not have the botName so we cannot throw a bot specific error here
		s.writeError(w, fmt.Sprintf("error in getBotState: %s\n", e))
		return
	}
	var req getBotStateRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeError(w, fmt.Sprintf("cannot have empty userID"))
		return
	}
	botName := req.BotName
	userData := req.UserData

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
