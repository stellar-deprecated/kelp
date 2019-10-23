package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/networking"
)

const kelpPrefsDirectory = "build"
const ccxtDownloadURL = "https://github.com/ccxt-rest/ccxt-rest/archive/v0.0.4.tar.gz"
const ccxtDownloadFilename = "ccxt-rest_v0.0.4.tar.gz"
const ccxtUntaredDirName = "ccxt-rest-0.0.4"
const ccxtBinOutputDir = "bin"

func main() {
	goosP := flag.String("goos", "", "GOOS for which to build")
	flag.Parse()
	goos := *goosP

	pkgos := ""
	if goos == "darwin" {
		pkgos = "macos"
	} else if goos == "linux" {
		pkgos = "linux"
	} else if goos == "windows" {
		pkgos = "win"
	} else {
		panic("unsupported goos flag: " + goos)
	}

	kos := kelpos.GetKelpOS()
	generateCcxtBinary(kos, pkgos)
}

func checkPkgTool(kos *kelpos.KelpOS) {
	fmt.Printf("checking for presence of `pkg` tool ...\n")
	_, e := kos.Blocking("pkg", "pkg -v")
	if e != nil {
		log.Fatal(errors.Wrap(e, "ensure that the `pkg` tool is installed correctly. You can get it from here: https://github.com/zeit/pkg"))
	}
	fmt.Printf("done\n")
}

func downloadCcxtSource(kos *kelpos.KelpOS, downloadDir string) {
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

func npmInstall(kos *kelpos.KelpOS, installDir string) {
	fmt.Printf("running npm install on directory %s ...\n", installDir)
	npmCmd := fmt.Sprintf("cd %s && npm install && cd -", installDir)
	_, e := kos.Blocking("npm", npmCmd)
	if e != nil {
		log.Fatal(errors.Wrap(e, "failed to run npm install"))
	}
	fmt.Printf("done\n")
}

// pkg --targets node8-macos-x64,node8-linux-x64,node8-win-x64 build/downloads/ccxt/ccxt-rest-0.0.4
func runPkgTool(kos *kelpos.KelpOS, sourceDir string, outDir string, pkgos string) {
	target := fmt.Sprintf("node8-%s-x64", pkgos)

	fmt.Printf("running pkg tool on source directory %s with output directory as %s on target platform %s ...\n", sourceDir, outDir, target)
	pkgCommand := fmt.Sprintf("pkg --out-path %s --targets %s %s", outDir, target, sourceDir)
	outputBytes, e := kos.Blocking("pkg", pkgCommand)
	if e != nil {
		log.Fatal(errors.Wrap(e, "failed to run pkg tool"))
	}
	fmt.Printf("done\n")

	copyDependencyFiles(kos, outDir, string(outputBytes))
}

func copyDependencyFiles(kos *kelpos.KelpOS, outDir string, pkgCmdOutput string) {
	fmt.Printf("copying dependency files to the output directory %s ...\n", outDir)
	for _, line := range strings.Split(pkgCmdOutput, "\n") {
		if !strings.Contains(line, "node_modules") {
			continue
		}
		filename := strings.TrimSpace(strings.ReplaceAll(line, "(MISSING)", ""))

		cpCmd := fmt.Sprintf("cp %s %s", filename, outDir)
		_, e := kos.Blocking("cp", cpCmd)
		if e != nil {
			log.Fatal(errors.Wrap(e, "failed to copy dependency file %s"+filename))
		}
	}
	fmt.Printf("done\n")
}

func generateCcxtBinary(kos *kelpos.KelpOS, pkgos string) {
	checkPkgTool(kos)

	downloadDir := filepath.Join(kelpPrefsDirectory, "downloads", "ccxt")
	sourceDir := filepath.Join(downloadDir, ccxtUntaredDirName)
	outDir := filepath.Join(downloadDir, ccxtBinOutputDir)

	downloadCcxtSource(kos, downloadDir)
	npmInstall(kos, sourceDir)
	runPkgTool(kos, sourceDir, outDir, pkgos)
}
