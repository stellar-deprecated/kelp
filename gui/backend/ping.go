package backend

import (
	"net/http"
)

func (s *APIServer) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Write([]byte("ok"))
}
