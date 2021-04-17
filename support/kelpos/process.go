package kelpos

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	"github.com/nikhilsaraf/go-tools/multithreading"
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
func (kos *KelpOS) SafeUnregister(userID string, namespace string) {
	kos.Unregister(userID, namespace)
}

// Stop unregisters and stops the command at the provided namespace
func (kos *KelpOS) Stop(userID string, namespace string) error {
	if p, exists := kos.GetProcess(userID, namespace); exists {
		e := kos.Unregister(userID, namespace)
		if e != nil {
			return fmt.Errorf("could not stop command because of an error when unregistering command for userID '%s' and namespace '%s': %s", userID, namespace, e)
		}

		log.Printf("killing process %d\n", p.Cmd.Process.Pid)
		return p.Cmd.Process.Kill()
	}
	return fmt.Errorf("process with userID '%s' and namespace '%s' does not exist", userID, namespace)
}

// Blocking runs a bash command and blocks
func (kos *KelpOS) Blocking(userID string, namespace string, cmd string) ([]byte, error) {
	p, e := kos.Background(userID, namespace, cmd)
	if e != nil {
		return nil, fmt.Errorf("could not run bash command in background '%s': %s", cmd, e)
	}

	// defer unregistration of process because regardless of whether it succeeds or fails it will not be active on the system anymore
	defer func() {
		eInner := kos.Unregister(userID, namespace)
		if eInner != nil {
			log.Fatalf("error unregistering bash command '%s': %s", cmd, eInner)
		}
	}()

	var outputBytes []byte
	var eRead error
	var eWait error
	threadTracker := multithreading.MakeThreadTracker()
	e = threadTracker.TriggerGoroutine(func(inputs []interface{}) {
		outputBytes, eRead = ioutil.ReadAll(p.Stdout)

		// wait for process to finish
		eWait = p.Cmd.Wait()
	}, nil)
	if e != nil {
		return nil, fmt.Errorf("error while triggering goroutine to read from process: %s", e)
	}
	// wait for threadTracker to finish -- we need to do this double-wait setup because p.Cmd.Wait() needs to be called after
	// we read but we still want to wait for this to happen because we are blocking on this command to finish.
	// see: https://github.com/hashicorp/go-plugin/issues/116#issuecomment-494153638
	threadTracker.Wait()

	// now check for errors
	if eWait != nil || eRead != nil {
		return nil, fmt.Errorf("error in bash command '%s' for userID '%s' and namespace '%s': (eWait=%s, outputBytes=%s, eRead=%v)",
			cmd, userID, namespace, eWait, string(outputBytes), eRead)
	}

	return outputBytes, nil
}

// Background runs the provided bash command in the background and registers the command
func (kos *KelpOS) Background(userID string, namespace string, cmd string) (*Process, error) {
	c := exec.Command("bash", "-c", cmd)
	// always execute commands from the working directory (specify as native since underlying OS handles it)
	// using dotKelpWorkingDir as working directory since all our config files and log files are located in here and we want
	// to have the shortest path lengths to accommodate for the 260 character file path limit in windows
	c.Dir = kos.dotKelpWorkingDir.Native()
	log.Printf("process.Background is executing command: '%s' from directory '%s'", c.String(), c.Dir)

	stdinWriter, e := c.StdinPipe()
	if e != nil {
		return nil, fmt.Errorf("could not get Stdin pipe for bash command '%s': %s", cmd, e)
	}
	stdoutReader, e := c.StdoutPipe()
	if e != nil {
		return nil, fmt.Errorf("could not get Stdout pipe for bash command '%s': %s", cmd, e)
	}

	e = c.Start()
	if e != nil {
		return nil, fmt.Errorf("could not start bash command '%s': %s", cmd, e)
	}

	p := &Process{
		Cmd:    c,
		Stdin:  stdinWriter,
		Stdout: stdoutReader,
	}
	e = kos.register(userID, namespace, p)
	if e != nil {
		return nil, fmt.Errorf("error registering bash command '%s': %s", cmd, e)
	}

	return p, nil
}

func (kos *KelpOS) register(userID string, namespace string, p *Process) error {
	kos.processLock.Lock()
	defer kos.processLock.Unlock()

	key := fmt.Sprintf("%s:%s", userID, namespace)
	if _, exists := kos.processes[key]; exists {
		return fmt.Errorf("process with key already exists: %s", key)
	}

	kos.processes[key] = *p
	if !kos.silentRegistrations {
		log.Printf("registered command under key '%s' with PID: %d, processes available: %v\n", key, p.Cmd.Process.Pid, kos.RegisteredProcesses())
	}
	return nil
}

// Unregister unregisters the command at the provided namespace, returning an error if needed
func (kos *KelpOS) Unregister(userID string, namespace string) error {
	kos.processLock.Lock()
	defer kos.processLock.Unlock()

	key := fmt.Sprintf("%s:%s", userID, namespace)
	if p, exists := kos.processes[key]; exists {
		delete(kos.processes, key)
		if !kos.silentRegistrations {
			log.Printf("unregistered command under key '%s' with PID: %d, processes available: %v\n", key, p.Cmd.Process.Pid, kos.RegisteredProcesses())
		}
		return nil
	}
	return fmt.Errorf("process with key does not exist: %s", key)
}

// GetProcess gets the process tied to the provided namespace
func (kos *KelpOS) GetProcess(userID string, namespace string) (*Process, bool) {
	kos.processLock.Lock()
	defer kos.processLock.Unlock()

	key := fmt.Sprintf("%s:%s", userID, namespace)
	p, exists := kos.processes[key]
	return &p, exists
}

// RegisteredProcesses returns the list of registered processes
func (kos *KelpOS) RegisteredProcesses() []string {
	list := []string{}
	for k := range kos.processes {
		list = append(list, k)
	}
	return list
}

// Mkdir function with a neat error message
func (kos *KelpOS) Mkdir(userID string, dirPath *OSPath) error {
	_, e := kos.Blocking(userID, "mkdir", fmt.Sprintf("mkdir -p %s", dirPath.Unix()))
	if e != nil {
		return fmt.Errorf("error running mkdir command for dir (%s): %s", dirPath.AsString(), e)
	}
	return nil
}
