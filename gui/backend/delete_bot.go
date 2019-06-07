package backend

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stellar/kelp/support/kelpos"
)

func (s *APIServer) deleteBot(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in deleteBot: %s\n", e))
		return
	}

	// only stop bot if current state is running
	botState, e := s.doGetBotState(botName)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in deleteBot unable to get botState: %s\n", e))
		return
	}
	log.Printf("current botState: %s\n", botState)
	if botState == kelpos.BotStateRunning {
		e = s.doStopBot(botName)
		if e != nil {
			s.writeError(w, fmt.Sprintf("error stopping bot when trying to delete: %s\n", e))
			return
		}
	}

	// unregister bot
	e = s.kos.Unregister(botName)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in deleteBot unable to unregister bot: %s\n", e))
		return
	}

	// TODO delete config

	w.WriteHeader(http.StatusOK)
}
