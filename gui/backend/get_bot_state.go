package backend

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func (s *APIServer) getBotState(w http.ResponseWriter, r *http.Request) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when reading request input: %s\n", e)))
	}
	botName := string(botNameBytes)

	b, e := s.kos.GetBot(botName)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error when getting bot state: %s\n", e)))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s\n", b.State)))
}
