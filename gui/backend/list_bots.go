package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/support/kelpos"
)

type listBotsRequest struct {
	UserData UserData `json:"user_data"`
}

func (s *APIServer) listBots(w http.ResponseWriter, r *http.Request) {
	log.Printf("listing bots\n")

	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req listBotsRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}

	bots, e := s.doListBots(req.UserData)
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

func (s *APIServer) doListBots(userData UserData) ([]model2.Bot, error) {
	bots := []model2.Bot{}
	resultBytes, e := s.kos.Blocking(userData.ID, "ls", fmt.Sprintf("ls %s | sort", s.botConfigsPathForUser(userData.ID).Unix()))
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

	ubd := s.kos.BotDataForUser(userData.toUser())
	for _, bot := range bots {
		botState, e := ubd.QueryBotState(bot.Name)
		if e != nil {
			return bots, fmt.Errorf("unable to query bot state for bot '%s': %s", bot.Name, e)
		}

		log.Printf("found bot '%s' with state '%s'\n", bot.Name, botState)
		// if page is reloaded then bot would already be registered, which is ok -- but we upsert here so it doesn't matter
		if botState != kelpos.InitState() {
			ubd.RegisterBotWithStateUpsert(&bot, botState)
		}
	}

	return bots, nil
}
