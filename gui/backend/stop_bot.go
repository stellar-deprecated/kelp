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

type stopBotRequest struct {
	UserData UserData `json:"user_data"`
	BotName  string   `json:"bot_name"`
}

func (s *APIServer) stopBot(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req stopBotRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}
	botName := req.BotName

	e = s.doStopBot(req.UserData, botName)
	if e != nil {
		s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
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

func (s *APIServer) doStopBot(userData UserData, botName string) error {
	e := s.kos.BotDataForUser(userData.toUser()).AdvanceBotState(botName, kelpos.BotStateRunning)
	if e != nil {
		return fmt.Errorf("error advancing bot state: %s", e)
	}

	e = s.kos.Stop(userData.ID, botName)
	if e != nil {
		return fmt.Errorf("error when killing bot %s: %s", botName, e)
	}
	log.Printf("stopped bot '%s'\n", botName)

	var numIterations uint8 = 1
	e = s.doStartBot(userData, botName, "delete", &numIterations, func() {
		eInner := s.deleteFinishCallback(userData, botName)
		if eInner != nil {
			s.addKelpErrorToMap(userData, makeKelpErrorResponseWrapper(
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

func (s *APIServer) deleteFinishCallback(userData UserData, botName string) error {
	log.Printf("deleted offers for bot '%s'\n", botName)

	e := s.kos.BotDataForUser(userData.toUser()).AdvanceBotState(botName, kelpos.BotStateStopping)
	if e != nil {
		return fmt.Errorf("error advancing bot state when manually attempting to stop bot: %s", e)
	}
	return nil
}
