package prefs

import (
	"fmt"
	"os"
)

// Preferences denotes a preferences file
type Preferences struct {
	filepath string
}

// Make creates a new Preferences struct
func Make(filepath string) *Preferences {
	return &Preferences{
		filepath: filepath,
	}
}

// FirstTime checks for the existence of the file
func (p *Preferences) FirstTime() bool {
	if _, e := os.Stat(p.filepath); os.IsNotExist(e) {
		return true
	}
	return false
}

// SetNotFirstTime saves a file on the file system to denote that this is not the first time
func (p *Preferences) SetNotFirstTime() error {
	emptyFile, e := os.Create(p.filepath)
	if e != nil {
		return fmt.Errorf("could not create file '%s' when setting not first time prefs file: %s", p.filepath, e)
	}
	emptyFile.Close()
	return nil
}
