package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	fs := http.Dir("./gui/web/build")
	e := vfsgen.Generate(fs, vfsgen.Options{
		Filename:        "./gui/filesystem_vfsdata_release.go",
		BuildTags:       "!debug",
		PackageName:     "gui",
		VariableName:    "FS",
		VariableComment: "file system for GUI as a blob",
	})
	if e != nil {
		log.Fatalln(e)
	}
}
