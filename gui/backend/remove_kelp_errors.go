package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// RemoveKelpErrorRequest is the outer object that contains the Kelp Error
type RemoveKelpErrorRequest struct {
	UserData     UserData `json:"user_data"`
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
	if strings.TrimSpace(kelpErrorRequest.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}
	userData := kelpErrorRequest.UserData

	// perform the actual action of removing the error
	removedMap := s.removeErrorsFromMap(userData, kelpErrorRequest.KelpErrorIDs)
	// reduce memory usage to prevent memory leaks in user data (since we remove entries here)
	s.removeKelpErrorUserDataIfEmpty(userData)

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

func (s *APIServer) removeErrorsFromMap(userData UserData, keIDs []string) (removedMap map[string]bool) {
	removedMap = map[string]bool{}

	kefu := s.kelpErrorsForUser(userData.ID)
	kefu.lock.Lock()
	defer kefu.lock.Unlock()

	for _, uuid := range keIDs {
		if _, exists := kefu.errorMap[uuid]; exists {
			delete(kefu.errorMap, uuid)
			removedMap[uuid] = true
		} else {
			removedMap[uuid] = false
		}
	}
	return removedMap
}
