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

	cmd, exists := s.kos.GetProcess(botName)
	if !exists {
		s.writeError(w, fmt.Sprintf("bot with name '%s' does not exist\n", botName))
		return
	}

	// send command
	writer, e := cmd.StdinPipe()
	if e != nil {
		s.writeError(w, fmt.Sprintf("unable to open stdin pipe of child kelp bot process (pid=%d) for writing: %s\n", cmd.Process.Pid, e))
		return
	}
	writer.Write([]byte("getBotInfo\n"))
	e = writer.Close()
	if e != nil {
		s.writeError(w, fmt.Sprintf("unable to close child kelp bot process's stdin writer (pid=%d): %s\n", cmd.Process.Pid, e))
		return
	}

	// read result
	reader, e := cmd.StdoutPipe()
	if e != nil {
		s.writeError(w, fmt.Sprintf("unable to open stdout pipe of child kelp bot process (pid=%d) for reading: %s\n", cmd.Process.Pid, e))
		return
	}
	outputBytes, e := ioutil.ReadAll(reader)
	if e != nil {
		s.writeError(w, fmt.Sprintf("unable to read output from stdout pipe of child kelp bot process (pid=%d): %s\n", cmd.Process.Pid, e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(outputBytes)
}
