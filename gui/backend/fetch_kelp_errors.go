package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// KelpErrorListResponseWrapper is the outer object that contains the Kelp Errors
type KelpErrorListResponseWrapper struct {
	KelpErrorList []KelpError `json:"kelp_error_list"`
}

type fetchKelpErrorsRequest struct {
	UserData UserData `json:"user_data"`
}

func (s *APIServer) fetchKelpErrors(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req fetchKelpErrorsRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}
	userData := req.UserData

	kefu := s.kelpErrorsForUser(userData.ID)
	kelpErrors := make([]KelpError, len(kefu.errorMap))
	i := 0
	for _, ke := range kefu.errorMap {
		kelpErrors[i] = ke
		i++
	}

	resp := KelpErrorListResponseWrapper{
		KelpErrorList: kelpErrors,
	}

	bytes, e := json.Marshal(resp)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.writeErrorJson(w, "unable to marshall kelp error list")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
