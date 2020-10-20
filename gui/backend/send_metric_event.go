package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type sendMetricEventRequest struct {
	EventName string `json:"name"`
	// TODO DS Populate request struct.
}

type sendMetricEventResponse struct {
	Success bool `json:"success"`
}

func (s *APIServer) sendMetricEvent(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s", e))
		return
	}
	log.Printf("sendMetricEvent requestJson: %s\n", string(bodyBytes))

	var req sendMetricEventRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}

	e = s.metricsTracker.SendGuiEvent(req.EventName)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error sending gui event %s: %s", req.EventName, e))
		return
	}

	s.writeJson(w, sendMetricEventResponse{Success: true})
}
