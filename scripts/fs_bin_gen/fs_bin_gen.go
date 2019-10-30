package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"

	"github.com/shurcooL/vfsgen"
)

const fsDev_filename = "./scripts/fs_bin_gen/gui/filesystem_vfsdata_dev.go"
const fs_filename = "./gui/filesystem_vfsdata.go"

func main() {
	envP := flag.String("env", "dev", "environment to use (dev / release)")
	flag.Parse()
	env := *envP

	if env == "dev" {
		generateWeb_Dev()
	} else if env == "release" {
		generateWeb_Release()
	} else {
		panic("unrecognized env flag: " + env)
	}
}

func generateWeb_Release() {
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

func generateWeb_Dev() {
	c := exec.Command("cp", fsDev_filename, fs_filename)
	e := c.Run()
	if e != nil {
		log.Fatalln(e)
	}
}
