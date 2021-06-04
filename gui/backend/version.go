package backend

import (
	"fmt"
	"net/http"
	"strings"
)

// this will be set automatically from root command
var versionString = ""

// SetVersionString sets the version string to be displayed in the GUI
func SetVersionString(guiVersion string, cliVersion string) {
	versionString = fmt.Sprintf("%s (%s)", strings.TrimSpace(guiVersion), strings.TrimSpace(cliVersion))
}

func (s *APIServer) version(w http.ResponseWriter, r *http.Request) {
	versionBytes := []byte(versionString)
	w.WriteHeader(http.StatusOK)
	w.Write(versionBytes)
}
