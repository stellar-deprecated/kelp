package backend

import (
	"fmt"
	"net/http"
)

func (s *APIServer) version(w http.ResponseWriter, r *http.Request) {
	resultBytes, e := s.runKelpCommandBlocking("version", "version | grep version | cut -d':' -f3")
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in version command: %s\n", e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultBytes)
}
