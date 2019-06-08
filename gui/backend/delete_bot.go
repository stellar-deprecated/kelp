package backend

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stellar/kelp/gui/model"
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

	for {
		botState, e := s.doGetBotState(botName)
		if e != nil {
			s.writeError(w, fmt.Sprintf("error in deleteBot for loop, unable to get botState: %s\n", e))
			return
		}
		log.Printf("deleteBot for loop, current botState: %s\n", botState)

		if botState == kelpos.BotStateStopped {
			break
		}

		time.Sleep(time.Second)
	}

	// unregister bot
	s.kos.SafeUnregisterBot(botName)

	// delete configs
	botPrefix := model.GetPrefix(botName)
	_, e = s.kos.Blocking("rm", fmt.Sprintf("rm %s/%s*", s.configsDir, botPrefix))
	if e != nil {
		s.writeError(w, fmt.Sprintf("error running rm command for bot configs: %s\n", e))
		return
	}
	log.Printf("removed bot configs for prefix '%s'\n", botPrefix)

	w.WriteHeader(http.StatusOK)
}
