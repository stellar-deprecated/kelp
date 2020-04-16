package kelpos

import (
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/stellar/kelp/gui/model2"
)

// KelpOS is a struct that manages all subprocesses started by this Kelp process
type KelpOS struct {
	workingBinDir       *OSPath
	processes           map[string]Process
	processLock         *sync.Mutex
	bots                map[string]*BotInstance
	botLock             *sync.Mutex
	silentRegistrations bool
}

// SetSilentRegistrations does not log every time we register and unregister commands
func (kos *KelpOS) SetSilentRegistrations() {
	kos.silentRegistrations = true
}

// Process contains all the pieces that can be used to control a given process
type Process struct {
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

// singleton is the singleton instance of KelpOS
var singleton *KelpOS

func init() {
	path, e := MakeOsPathBase()
	if e != nil {
		panic(e)
	}

	singleton = &KelpOS{
		workingBinDir: path,
		processes:     map[string]Process{},
		processLock:   &sync.Mutex{},
		bots:          map[string]*BotInstance{},
		botLock:       &sync.Mutex{},
	}
}

// BotInstance is an instance of a given bot along with the metadata
type BotInstance struct {
	Bot   *model2.Bot
	State BotState
}

// GetKelpOS gets the singleton instance
func GetKelpOS() *KelpOS {
	if singleton == nil {
		panic(fmt.Errorf("there is a cycle stemming from the init() method since singleton was nil"))
	}
	return singleton
}
