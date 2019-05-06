package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// APIServer is an instance of the API service
type APIServer struct {
	binPath string
}

// MakeAPIServer is a factory method
func MakeAPIServer() (*APIServer, error) {
	binPath, e := filepath.Abs(os.Args[0])
	if e != nil {
		return nil, fmt.Errorf("could not get binPath of currently running binary: %s", e)
	}

	return &APIServer{
		binPath: binPath,
	}, nil
}

func (s *APIServer) runCommand(cmd string) ([]byte, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	bytes, e := exec.Command("bash", "-c", cmdString).Output()
	if e != nil {
		return nil, fmt.Errorf("could not run bash command '%s': %s", cmd, e)
	}
	return bytes, nil
}
