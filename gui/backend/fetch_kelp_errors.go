package backend

import (
	"encoding/json"
	"net/http"
)

// KelpErrorListResponseWrapper is the outer object that contains the Kelp Errors
type KelpErrorListResponseWrapper struct {
	KelpErrorList []KelpError `json:"kelp_error_list"`
}

func (s *APIServer) fetchKelpErrors(w http.ResponseWriter, r *http.Request) {
	kelpErrors := make([]KelpError, len(s.kelpErrorMap))
	i := 0
	for _, ke := range s.kelpErrorMap {
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
