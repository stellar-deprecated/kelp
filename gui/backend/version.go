package backend

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *APIServer) version(w http.ResponseWriter, r *http.Request) {
	guiVersionBytes, e := s.runKelpCommandBlocking("version", "version | grep 'gui version' | cut -d':' -f2,3")
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in version command: %s\n", e))
		return
	}
	cliVersionBytes, e := s.runKelpCommandBlocking("version", "version | grep 'cli version' | cut -d':' -f2,3")
	if e != nil {
		s.writeError(w, fmt.Sprintf("error in version command: %s\n", e))
		return
	}

	versionBytes := []byte(fmt.Sprintf("%s (%s)", strings.TrimSpace(string(guiVersionBytes)), strings.TrimSpace(string(cliVersionBytes))))
	w.WriteHeader(http.StatusOK)
	w.Write(versionBytes)
}
