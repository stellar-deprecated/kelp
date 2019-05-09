package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stellar/kelp/support/utils"
)

// APIServer is an instance of the API service
type APIServer struct {
	dirPath    string
	binPath    string
	configsDir string
	logsDir    string
}

// MakeAPIServer is a factory method
func MakeAPIServer() (*APIServer, error) {
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
	}, nil
}

func (s *APIServer) runKelpCommand(cmd string) ([]byte, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return runBashCommand(cmdString)
}

func (s *APIServer) runKelpCommandStreaming(cmd string) error {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return utils.RunCommandStreamOutput(exec.Command("bash", "-c", cmdString))
}

func runBashCommand(cmd string) ([]byte, error) {
	resultBytes, e := exec.Command("bash", "-c", cmd).Output()
	if e != nil {
		return nil, fmt.Errorf("could not run bash command '%s': %s", cmd, e)
	}
	return resultBytes, nil
}
