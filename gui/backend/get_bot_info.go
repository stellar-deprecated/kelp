package backend

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/kelp/support/utils"
)

func (s *APIServer) getBotInfo(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error parsing bot name in getBotInfo: %s\n", e))
		return
	}

	p, exists := s.kos.GetProcess(botName)
	if !exists {
		log.Printf("kelp bot process with name '%s' does not exist; processes available: %v\n", botName, s.kos.RegisteredProcesses())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{}"))
		return
	}

	log.Printf("getBotInfo is making IPC request for botName: %s\n", botName)
	p.PipeIn.Write([]byte("getBotInfo\n"))
	scanner := bufio.NewScanner(p.PipeOut)
	output := ""
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, utils.IPCBoundary) {
			break
		}
		output += text
	}
	var buf bytes.Buffer
	e = json.Indent(&buf, []byte(output), "", "  ")
	if e != nil {
		log.Printf("cannot indent json response (error=%s), json_response: %s\n", e, output)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{}"))
		return
	}
	log.Printf("getBotInfo returned IPC response for botName '%s': %s\n", botName, buf.String())

	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
