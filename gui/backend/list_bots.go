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

	bots, e := s.doListBots()
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error encountered while listing bots: %s", e))
		return
	}

	// serialize and return
	botsJSON, e := json.Marshal(bots)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to serialize bots: %s\n", e))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(botsJSON)
}

func (s *APIServer) doListBots() ([]model2.Bot, error) {
	bots := []model2.Bot{}
	resultBytes, e := s.kos.Blocking("ls", fmt.Sprintf("ls %s | sort", s.botConfigsPath.Unix()))
	if e != nil {
		return bots, fmt.Errorf("error when listing bots: %s", e)
	}
	configFiles := string(resultBytes)
	files := strings.Split(configFiles, "\n")

	// run till one less than length of files because the last name will end in a newline
	for i := 0; i < len(files)-1; i += 2 {
		bot := model2.FromFilenames(files[i+1], files[i])
		bots = append(bots, *bot)
	}
	log.Printf("bots available: %v", bots)

	for _, bot := range bots {
		botState, e := s.kos.QueryBotState(bot.Name)
		if e != nil {
			return bots, fmt.Errorf("unable to query bot state for bot '%s': %s", bot.Name, e)
		}

		log.Printf("found bot '%s' with state '%s'\n", bot.Name, botState)
		// if page is reloaded then bot would already be registered, which is ok -- but we upsert here so it doesn't matter
		if botState != kelpos.InitState() {
			s.kos.RegisterBotWithStateUpsert(&bot, botState)
		}
	}

	return bots, nil
}
