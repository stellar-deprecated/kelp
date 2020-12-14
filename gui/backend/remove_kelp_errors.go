package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// RemoveKelpErrorRequest is the outer object that contains the Kelp Error
type RemoveKelpErrorRequest struct {
	KelpErrorIDs []string `json:"kelp_error_ids"`
}

// RemoveKelpErrorResponse is the outer object that contains the Kelp Error
type RemoveKelpErrorResponse struct {
	RemovedMap map[string]bool `json:"removed_map"`
}

func (s *APIServer) removeKelpErrors(w http.ResponseWriter, r *http.Request) {
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
	removedMap := s.removeErrorsFromMap(kelpErrorRequest.KelpErrorIDs)

	resp := RemoveKelpErrorResponse{RemovedMap: removedMap}
	bytes, e := json.Marshal(resp)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.writeErrorJson(w, fmt.Sprintf("unable to marshall kelp error response: %+v", resp))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func (s *APIServer) removeErrorsFromMap(keIDs []string) (removedMap map[string]bool) {
	removedMap = map[string]bool{}

	s.kelpErrorMapLock.Lock()
	defer s.kelpErrorMapLock.Unlock()

	for _, uuid := range keIDs {
		if _, exists := s.kelpErrorMap[uuid]; exists {
			delete(s.kelpErrorMap, uuid)
			removedMap[uuid] = true
		} else {
			removedMap[uuid] = false
		}
	}
	return removedMap
}
