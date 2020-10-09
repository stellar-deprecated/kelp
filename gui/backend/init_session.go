package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type initSessionRequest struct {
	// TODO: Define initSession request fields.
}

type initSessionResponse struct {
	Success bool `json:"success"`
}

func (s *APIServer) initSession(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s", e))
		return
	}
	log.Printf("initSession requestJson: %s\n", string(bodyBytes))

	var req initSessionRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}

	// TODO: Process appropriate fields in request body.

	// TODO: Send appropriate request to `s.metricsTracker`.
	e = s.metricsTracker.SendStartupEvent()
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error sending startup event: %s", e))
	}

	s.writeJson(w, initSessionResponse{Success: true})
}
