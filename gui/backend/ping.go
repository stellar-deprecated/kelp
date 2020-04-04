package backend

import (
	"net/http"
)

func (s *APIServer) ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("ok"))
}
