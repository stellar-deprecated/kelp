package kelpos

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"os/user"
	"sync"

	"github.com/stellar/go/support/errors"
	"github.com/stellar/kelp/gui/model2"
)

const dotKelpDir = ".kelp"

// KelpOS is a struct that manages all subprocesses started by this Kelp process
type KelpOS struct {
	binDir              *OSPath
	dotKelpWorkingDir   *OSPath
	processes           map[string]Process
	processLock         *sync.Mutex
	bots                map[string]*BotInstance
	botLock             *sync.Mutex
	silentRegistrations bool
}

// GetBinDir accessor
func (kos *KelpOS) GetBinDir() *OSPath {
	return kos.binDir
}

// GetDotKelpWorkingDir accessor
func (kos *KelpOS) GetDotKelpWorkingDir() *OSPath {
	return kos.dotKelpWorkingDir
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
	binDir, e := MakeOsPathBase()
	if e != nil {
		panic(errors.Wrap(e, "could not make binDir"))
	}
	log.Printf("binDir initialized: %s", binDir.AsString())

	usr, e := user.Current()
	if e != nil {
		panic(errors.Wrap(e, "could not fetch current user (need to get home directory)"))
	}
	usrHomeDir, e := binDir.MakeFromNativePath(usr.HomeDir)
	if e != nil {
		panic(errors.Wrap(e, "could not make usrHomeDir from usr.HomeDir="+usr.HomeDir))
	}
	log.Printf("Kelp is being run from user '%s' (Uid=%s, Name=%s, HomeDir=%s)", usr.Username, usr.Uid, usr.Name, usrHomeDir.AsString())

	// file path for windows needs to be 260 characters (https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file)
	// so we want to put it closer to the root volume in ~/.kelp (or C:\.kelp) so it does not throw an error
	dotKelpWorkingDir := usrHomeDir.Join(dotKelpDir)
	log.Printf("dotKelpWorkingDir initialized: %s", dotKelpWorkingDir.AsString())
	// manually make dotKelpWorkingDir so we can use it as the working dir for kelpos
	mkDotKelpWorkingDir := fmt.Sprintf("mkdir -p %s", dotKelpWorkingDir.Unix())
	e = exec.Command("bash", "-c", mkDotKelpWorkingDir).Run()
	if e != nil {
		panic(fmt.Errorf("could not run raw command 'bash -c %s': %s", mkDotKelpWorkingDir, e))
	}

	// using dotKelpWorkingDir as working directory since all our config files and log files are located in here and we want
	// to have the shortest path lengths to accommodate for the 260 character file path limit in windows
	singleton = &KelpOS{
		binDir:            binDir,
		dotKelpWorkingDir: dotKelpWorkingDir,
		processes:         map[string]Process{},
		processLock:       &sync.Mutex{},
		bots:              map[string]*BotInstance{},
		botLock:           &sync.Mutex{},
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
