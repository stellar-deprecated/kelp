package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/support/kelpos"
)

type deleteBotRequest struct {
	UserData UserData `json:"user_data"`
	BotName  string   `json:"bot_name"`
}

func (s *APIServer) deleteBot(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req deleteBotRequest
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

	// only stop bot if current state is running
	botState, e := s.doGetBotState(req.UserData, botName)
	if e != nil {
		s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelWarning,
			fmt.Sprintf("unable to get botState: %s\n", e),
		))
		return
	}
	log.Printf("current botState: %s\n", botState)
	if botState == kelpos.BotStateRunning {
		e = s.doStopBot(req.UserData, botName)
		if e != nil {
			s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelWarning,
				fmt.Sprintf("could not stop bot when trying to delete: %s\n", e),
			))
			return
		}
	}

	for {
		botState, e := s.doGetBotState(req.UserData, botName)
		if e != nil {
			s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelError,
				fmt.Sprintf("unable to get botState: %s\n", e),
			))
			return
		}
		log.Printf("deleteBot for loop, current botState: %s\n", botState)

		if botState == kelpos.BotStateStopped || botState == kelpos.BotStateInitializing {
			break
		}

		time.Sleep(time.Second)
	}

	// unregister bot
	s.kos.BotDataForUser(req.UserData.toUser()).SafeUnregisterBot(botName)

	// delete configs
	botPrefix := model2.GetPrefix(botName)
	botConfigPath := s.botConfigsPathForUser(req.UserData.ID).Join(botPrefix)
	_, e = s.kos.Blocking(req.UserData.ID, "rm", fmt.Sprintf("rm %s*", botConfigPath.Unix()))
	if e != nil {
		s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("could not run rm command for bot configs: %s\n", e),
		))
		return
	}
	log.Printf("removed bot configs for prefix '%s'\n", botPrefix)

	w.WriteHeader(http.StatusOK)
}
