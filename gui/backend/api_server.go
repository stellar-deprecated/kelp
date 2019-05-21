package backend

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/stellar/kelp/support/utils"
)

// APIServer is an instance of the API service
type APIServer struct {
	dirPath     string
	binPath     string
	configsDir  string
	logsDir     string
	processes   map[string]*exec.Cmd
	processLock *sync.Mutex
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
		dirPath:     dirPath,
		binPath:     binPath,
		configsDir:  configsDir,
		logsDir:     logsDir,
		processes:   map[string]*exec.Cmd{},
		processLock: &sync.Mutex{},
	}, nil
}

func (s *APIServer) registerCommand(namespace string, c *exec.Cmd) error {
	s.processLock.Lock()
	defer s.processLock.Unlock()

	if _, exists := s.processes[namespace]; exists {
		return fmt.Errorf("process with namespace already exists: %s", namespace)
	}

	s.processes[namespace] = c
	log.Printf("registered command under namespace '%s' with PID: %d", namespace, c.Process.Pid)
	return nil
}

func (s *APIServer) unregisterCommand(namespace string) error {
	s.processLock.Lock()
	defer s.processLock.Unlock()

	if c, exists := s.processes[namespace]; exists {
		delete(s.processes, namespace)
		log.Printf("unregistered command under namespace '%s' with PID: %d", namespace, c.Process.Pid)
		return nil
	}
	return fmt.Errorf("process with namespace does not exist: %s", namespace)
}

func (s *APIServer) getCommand(namespace string) (*exec.Cmd, bool) {
	s.processLock.Lock()
	defer s.processLock.Unlock()

	c, exists := s.processes[namespace]
	return c, exists
}

func (s *APIServer) safeUnregisterCommand(namespace string) {
	s.unregisterCommand(namespace)
}

func (s *APIServer) stopCommand(namespace string) error {
	if c, exists := s.getCommand(namespace); exists {
		e := s.unregisterCommand(namespace)
		if e != nil {
			return fmt.Errorf("could not stop command because of an error when unregistering command for namespace '%s': %s", namespace, e)
		}

		log.Printf("killing process %d\n", c.Process.Pid)
		return c.Process.Kill()
	}
	return fmt.Errorf("process with namespace does not exist: %s", namespace)
}

func (s *APIServer) runKelpCommandBlocking(namespace string, cmd string) ([]byte, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return s.runBashCommandBlocking(namespace, cmdString)
}

func (s *APIServer) runKelpCommandBackground(namespace string, cmd string) (*exec.Cmd, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return s.runBashCommandBackground(namespace, cmdString, nil)
}

func (s *APIServer) runKelpCommandStreaming(cmd string) error {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return utils.RunCommandStreamOutput(exec.Command("bash", "-c", cmdString))
}

func (s *APIServer) runBashCommandBlocking(namespace string, cmd string) ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	c, e := s.runBashCommandBackground(namespace, cmd, writer)
	if e != nil {
		return nil, fmt.Errorf("could not run bash command in background '%s': %s", cmd, e)
	}

	e = c.Wait()
	if e != nil {
		return nil, fmt.Errorf("error waiting for bash command '%s': %s", cmd, e)
	}

	e = s.unregisterCommand(namespace)
	if e != nil {
		return nil, fmt.Errorf("error unregistering bash command '%s': %s", cmd, e)
	}

	return buf.Bytes(), nil
}

func (s *APIServer) runBashCommandBackground(namespace string, cmd string, writer io.Writer) (*exec.Cmd, error) {
	c := exec.Command("bash", "-c", cmd)
	if writer != nil {
		c.Stdout = writer
	}

	e := c.Start()
	if e != nil {
		return c, fmt.Errorf("could not start bash command '%s': %s", cmd, e)
	}

	e = s.registerCommand(namespace, c)
	if e != nil {
		return nil, fmt.Errorf("error registering bash command '%s': %s", cmd, e)
	}

	return c, nil
}
