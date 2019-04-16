// +build dev

package gui

import (
	"net/http"
	"path/filepath"
)

var guiBuildDir = filepath.Join("gui", "web", "build")

// file system for GUI
var FS = http.Dir(guiBuildDir)
