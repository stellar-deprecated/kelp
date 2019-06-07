package backend

import (
	"fmt"
	"net/http"
)

func (s *APIServer) deleteBot(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in deleteBot: %s\n", e))
		return
	}

	// TODO only stop bot if current state is running

	e = s.doStopBot(botName)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error stopping bot when trying to delete: %s\n", e))
		return
	}

	// TODO unregister bot and delete config

	w.WriteHeader(http.StatusOK)
}
