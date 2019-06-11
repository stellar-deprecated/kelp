package kelpos

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/stellar/kelp/gui/model"
)

// KelpOS is a struct that manages all subprocesses started by this Kelp process
type KelpOS struct {
	processes   map[string]Process
	processLock *sync.Mutex
	bots        map[string]*BotInstance
	botLock     *sync.Mutex
}

// Process contains all the pieces that can be used to control a given process
type Process struct {
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	PipeIn  *os.File
	PipeOut *os.File
}

// singleton is the singleton instance of KelpOS
var singleton *KelpOS

func init() {
	singleton = &KelpOS{
		processes:   map[string]Process{},
		processLock: &sync.Mutex{},
		bots:        map[string]*BotInstance{},
		botLock:     &sync.Mutex{},
	}
}

// BotInstance is an instance of a given bot along with the metadata
type BotInstance struct {
	Bot   *model.Bot
	State BotState
}

// GetKelpOS gets the singleton instance
func GetKelpOS() *KelpOS {
	return singleton
}
