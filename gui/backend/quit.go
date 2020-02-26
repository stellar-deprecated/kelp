package backend

import (
	"net/http"
	"time"
)

func (s *APIServer) quit(w http.ResponseWriter, r *http.Request) {
	go func() {
		// sleep so we can respond to the request
		time.Sleep(1 * time.Second)
		s.quitFn()
	}()
	w.WriteHeader(http.StatusOK)
}
