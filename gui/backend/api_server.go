package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stellar/kelp/support/kelpos"
)

// APIServer is an instance of the API service
type APIServer struct {
	dirPath    string
	binPath    string
	configsDir string
	logsDir    string
	kos        *kelpos.KelpOS
}

// MakeAPIServer is a factory method
func MakeAPIServer(kos *kelpos.KelpOS) (*APIServer, error) {
	binPath, e := filepath.Abs(os.Args[0])
	if e != nil {
		return nil, fmt.Errorf("could not get binPath of currently running binary: %s", e)
	}

	dirPath := filepath.Dir(binPath)
	configsDir := dirPath + "/ops/configs"
	logsDir := dirPath + "/ops/logs"

	return &APIServer{
		dirPath:    dirPath,
		binPath:    binPath,
		configsDir: configsDir,
		logsDir:    logsDir,
		kos:        kos,
	}, nil
}

func (s *APIServer) parseBotName(r *http.Request) (string, error) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return "", fmt.Errorf("error when reading request input: %s\n", e)
	}
	return string(botNameBytes), nil
}

func (s *APIServer) writeError(w http.ResponseWriter, message string) {
	log.Print(message)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
}

func (s *APIServer) runKelpCommandBlocking(namespace string, cmd string) ([]byte, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return s.kos.Blocking(namespace, cmdString)
}

func (s *APIServer) runKelpCommandBackground(namespace string, cmd string) (*exec.Cmd, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return s.kos.Background(namespace, cmdString, nil)
}
