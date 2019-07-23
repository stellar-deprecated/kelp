package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/support/kelpos"
)

func (s *APIServer) listBots(w http.ResponseWriter, r *http.Request) {
	log.Printf("listing bots\n")
	resultBytes, e := s.kos.Blocking("ls", fmt.Sprintf("ls %s | sort", s.configsDir))
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when listing bots: %s\n", e))
		return
	}
	configFiles := string(resultBytes)
	files := strings.Split(configFiles, "\n")

	bots := []model2.Bot{}
	// run till one less than length of files because the last name will end in a newline
	for i := 0; i < len(files)-1; i += 2 {
		bot := model2.FromFilenames(files[i+1], files[i])
		bots = append(bots, *bot)
	}
	log.Printf("bots available: %v", bots)

	for _, bot := range bots {
		botState, e := s.kos.QueryBotState(bot.Name)
		if e != nil {
			s.writeErrorJson(w, fmt.Sprintf("unable to query bot state for bot '%s': %s\n", bot.Name, e))
			return
		}

		log.Printf("found bot '%s' with state '%s'\n", bot.Name, botState)
		// if page is reloaded then bot would already be registered, which is ok -- but we upsert here so it doesn't matter
		if botState != kelpos.InitState() {
			s.kos.RegisterBotWithStateUpsert(&bot, botState)
		}
	}

	// serialize and return
	botsJson, e := json.Marshal(bots)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to serialize bots: %s\n", e))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(botsJson)
}
