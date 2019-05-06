package backend

import (
	"fmt"
	"os/exec"
)

func runCommand(cmd string) ([]byte, error) {
	b, e := exec.Command("bash", "-c", cmd).Output()
	if e != nil {
		return nil, fmt.Errorf("could not run bash command '%s': %s", cmd, e)
	}
	return b, nil
}
