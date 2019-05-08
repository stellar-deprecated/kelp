package backend

import (
	"fmt"
	"net/http"
)

func (s *APIServer) version(w http.ResponseWriter, r *http.Request) {
	bytes, e := s.runKelpCommand("version | grep version | cut -d':' -f3")
	if e != nil {
		w.Write([]byte(fmt.Sprintf("error in version command: %s\n", e)))
		return
	}
	w.Write(bytes)
}
