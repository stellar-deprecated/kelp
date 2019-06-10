package kelpos

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
)

// StreamOutput runs the provided command in a streaming fashion
func (kos *KelpOS) StreamOutput(command *exec.Cmd) error {
	stdout, e := command.StdoutPipe()
	if e != nil {
		return fmt.Errorf("error while creating Stdout pipe: %s", e)
	}
	command.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("\t%s\n", line)
	}

	e = command.Wait()
	if e != nil {
		return fmt.Errorf("could not execute command: %s", e)
	}
	return nil
}

// SafeUnregister ignores erros when unregistering the command at the provided namespace
func (kos *KelpOS) SafeUnregister(namespace string) {
	kos.Unregister(namespace)
}

// Stop unregisters and stops the command at the provided namespace
func (kos *KelpOS) Stop(namespace string) error {
	if p, exists := kos.GetProcess(namespace); exists {
		e := kos.Unregister(namespace)
		if e != nil {
			return fmt.Errorf("could not stop command because of an error when unregistering command for namespace '%s': %s", namespace, e)
		}

		log.Printf("killing process %d\n", p.Cmd.Process.Pid)
		return p.Cmd.Process.Kill()
	}
	return fmt.Errorf("process with namespace does not exist: %s", namespace)
}

// Blocking runs a bash command and blocks
func (kos *KelpOS) Blocking(namespace string, cmd string) ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	c, e := kos.Background(namespace, cmd, writer)
	if e != nil {
		return nil, fmt.Errorf("could not run bash command in background '%s': %s", cmd, e)
	}

	e = c.Wait()
	if e != nil {
		return nil, fmt.Errorf("error waiting for bash command '%s': %s", cmd, e)
	}

	e = kos.Unregister(namespace)
	if e != nil {
		return nil, fmt.Errorf("error unregistering bash command '%s': %s", cmd, e)
	}

	return buf.Bytes(), nil
}

// Background runs the provided bash command in the background and registers the command
func (kos *KelpOS) Background(namespace string, cmd string, writer io.Writer) (*exec.Cmd, error) {
	c := exec.Command("bash", "-c", cmd)
	if writer != nil {
		c.Stdout = writer
	}

	stdinWriter, e := c.StdinPipe()
	if e != nil {
		return c, fmt.Errorf("could not get Stdin pipe for bash command '%s': %s", cmd, e)
	}
	stdoutReader, e := c.StdoutPipe()
	if e != nil {
		return c, fmt.Errorf("could not get Stdout pipe for bash command '%s': %s", cmd, e)
	}

	e = c.Start()
	if e != nil {
		return c, fmt.Errorf("could not start bash command '%s': %s", cmd, e)
	}

	e = kos.register(namespace, c, stdinWriter, stdoutReader)
	if e != nil {
		return nil, fmt.Errorf("error registering bash command '%s': %s", cmd, e)
	}

	return c, nil
}

func (kos *KelpOS) register(namespace string, c *exec.Cmd, stdinWriter io.WriteCloser, stdoutReader io.ReadCloser) error {
	kos.processLock.Lock()
	defer kos.processLock.Unlock()

	if _, exists := kos.processes[namespace]; exists {
		return fmt.Errorf("process with namespace already exists: %s", namespace)
	}

	kos.processes[namespace] = Process{
		Cmd:    c,
		Stdin:  stdinWriter,
		Stdout: stdoutReader,
	}
	log.Printf("registered command under namespace '%s' with PID: %d", namespace, c.Process.Pid)
	return nil
}

// Unregister unregisters the command at the provided namespace, returning an error if needed
func (kos *KelpOS) Unregister(namespace string) error {
	kos.processLock.Lock()
	defer kos.processLock.Unlock()

	if p, exists := kos.processes[namespace]; exists {
		delete(kos.processes, namespace)
		log.Printf("unregistered command under namespace '%s' with PID: %d", namespace, p.Cmd.Process.Pid)
		return nil
	}
	return fmt.Errorf("process with namespace does not exist: %s", namespace)
}

// GetProcess gets the process tied to the provided namespace
func (kos *KelpOS) GetProcess(namespace string) (*Process, bool) {
	kos.processLock.Lock()
	defer kos.processLock.Unlock()

	p, exists := kos.processes[namespace]
	return &p, exists
}
