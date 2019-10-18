package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"

	"github.com/shurcooL/vfsgen"
)

const fsDev_filename = "./scripts/fs_gen/filesystem_vfsdata_dev.go"
const fs_filename = "./gui/filesystem_vfsdata.go"

func main() {
	envP := flag.String("env", "dev", "environment to use (dev / release)")
	flag.Parse()
	env := *envP

	if env == "dev" {
		generateDev()
	} else if env == "release" {
		generateRelease()
	} else {
		panic("unrecognized env flag: " + env)
	}
}

func generateRelease() {
	fs := http.Dir("./gui/web/build")
	e := vfsgen.Generate(fs, vfsgen.Options{
		Filename:        fs_filename,
		PackageName:     "gui",
		VariableName:    "FS",
		VariableComment: "file system for GUI as a blob",
	})
	if e != nil {
		log.Fatalln(e)
	}
}

func generateDev() {
	c := exec.Command("cp", fsDev_filename, fs_filename)
	e := c.Run()
	if e != nil {
		log.Fatalln(e)
	}
}
