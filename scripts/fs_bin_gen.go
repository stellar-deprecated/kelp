package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stellar/kelp/support/networking"

	"github.com/shurcooL/vfsgen"
	"github.com/stellar/kelp/support/kelpos"
)

const kelpPrefsDirectory = "build"
const fsDev_filename = "./scripts/fs_gen/filesystem_vfsdata_dev.go"
const fs_filename = "./gui/filesystem_vfsdata.go"
const ccxtDownloadURL = "https://github.com/ccxt-rest/ccxt-rest/archive/v0.0.4.tar.gz"
const ccxtDownloadFilename = "ccxt-rest_v0.0.4.tar.gz"
const ccxtUntaredDirName = "ccxt-rest-0.0.4"

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

	kos := kelpos.GetKelpOS()
	generateCcxtBinary(kos)
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

func checkPkgTool(kos *kelpos.KelpOS) {
	fmt.Printf("checking for presence of `pkg` tool ...\n")
	_, e := kos.Blocking("pkg", "pkg -v")
	if e != nil {
		log.Fatal("ensure that the `pkg` tool is installed correctly. You can get it from here: https://github.com/zeit/pkg")
	}
	fmt.Printf("done\n")
}

func downloadCcxtSource(kos *kelpos.KelpOS) {
	downloadDir := filepath.Join(kelpPrefsDirectory, "downloads", "ccxt")
	fmt.Printf("making directory where we can download ccxt file %s ...\n", downloadDir)
	e := kos.Mkdir(downloadDir)
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not make directory for downloadDir "+downloadDir))
	}
	fmt.Printf("done\n")

	fmt.Printf("downloading file from URL %s ...\n", ccxtDownloadURL)
	downloadFilePath := filepath.Join(downloadDir, ccxtDownloadFilename)
	e = networking.DownloadFile(ccxtDownloadURL, downloadFilePath)
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not download ccxt tar.gz file"))
	}
	fmt.Printf("done\n")

	fmt.Printf("untaring file %s ...\n", downloadFilePath)
	_, e = kos.Blocking("tar", fmt.Sprintf("tar xvf %s -C %s", downloadFilePath, downloadDir))
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not untar ccxt file"))
	}
	fmt.Printf("done\n")
}

func generateCcxtBinary(kos *kelpos.KelpOS) {
	checkPkgTool(kos)
	downloadCcxtSource(kos)
}
