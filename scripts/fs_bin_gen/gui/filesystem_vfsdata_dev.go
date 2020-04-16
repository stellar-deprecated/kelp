package gui

import (
	"net/http"
	"path/filepath"
)

// build dir is gui/web/build because we run from the root kelp directory in dev mode
var guiBuildDir = filepath.Join("gui", "web", "build")

// file system for GUI
var FS = http.Dir(guiBuildDir)
