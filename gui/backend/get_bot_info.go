package backend

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func (s *APIServer) getBotInfo(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error parsing bot name in getBotInfo: %s\n", e))
		return
	}

	p, exists := s.kos.GetProcess(botName)
	if !exists {
		s.writeError(w, fmt.Sprintf("bot with name '%s' does not exist\n", botName))
		return
	}

	// make request via IPC
	p.Stdin.Write([]byte("getBotInfo\n"))
	outputBytes, e := ioutil.ReadAll(p.Stdout)
	if e != nil {
		s.writeError(w, fmt.Sprintf("unable to read output from stdout pipe of child kelp bot process (pid=%d): %s\n", p.Cmd.Process.Pid, e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(outputBytes)
}
