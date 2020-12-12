package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// RemoveKelpErrorRequest is the outer object that contains the Kelp Error
type RemoveKelpErrorRequest struct {
	KelpError KelpError `json:"kelp_error"`
}

// RemoveKelpErrorResponse is the outer object that contains the Kelp Error
type RemoveKelpErrorResponse struct {
	Removed bool `json:"removed"`
}

func (s *APIServer) removeKelpError(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request body input: %s", e))
		return
	}

	var kelpErrorRequest RemoveKelpErrorRequest
	e = json.Unmarshal(bodyBytes, &kelpErrorRequest)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to parse kelp error input from request as json: %s", e))
		return
	}

	// perform the actual action of removing the error
	removed := s.removeErrorFromMap(kelpErrorRequest.KelpError)

	resp := RemoveKelpErrorResponse{Removed: removed}
	bytes, e := json.Marshal(resp)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.writeErrorJson(w, fmt.Sprintf("unable to marshall kelp error response: %+v", resp))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (s *APIServer) removeErrorFromMap(ke KelpError) bool {
	key := ke.String()

	s.kelpErrorMapLock.Lock()
	defer s.kelpErrorMapLock.Unlock()

	if _, exists := s.kelpErrorMap[key]; exists {
		delete(s.kelpErrorMap, key)
		return true
	}
	return false
}
