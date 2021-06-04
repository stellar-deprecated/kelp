package backend

import (
	"fmt"
	"net/http"
	"time"
)

func (s *APIServer) quit(w http.ResponseWriter, r *http.Request) {
	if s.enableKaas {
		w.WriteHeader(http.StatusInternalServerError)
		panic(fmt.Errorf("quit functionality should have been disabled in routes when running in KaaS mode"))
	}

	go func() {
		// sleep so we can respond to the request
		time.Sleep(1 * time.Second)
		s.quitFn()
	}()
	w.WriteHeader(http.StatusOK)
}
