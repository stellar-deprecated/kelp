package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
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
const nodeVersionMatchRegex = "v8.[0-9]+.[0-9]+"

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
	kos.SetSilentRegistrations()

	zipFoldername := fmt.Sprintf("ccxt-rest_%s-x64", goos)
	// no need to pass a userID since we are not running under the context of any user at this point
	generateCcxtBinary(kos, "_", pkgos, zipFoldername)
}

func checkNodeVersion(kos *kelpos.KelpOS, userID string) {
	fmt.Printf("checking node version ... ")

	version, e := kos.Blocking(userID, "node", "node -v")
	if e != nil {
		log.Fatal(errors.Wrap(e, "ensure that the `pkg` tool is installed correctly. You can get it from here https://github.com/zeit/pkg or by running `npm install -g pkg`"))
	}

	match, e := regexp.Match(nodeVersionMatchRegex, version)
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not match regex against node version"))
	}
	if !match {
		log.Fatal("node version will fail to compile a successful binary because of the requirements of ccxt-rest's dependencies, should use v8.x.x instead of " + string(version))
	}

	fmt.Printf("valid\n")
}

func checkPkgTool(kos *kelpos.KelpOS, userID string) {
	fmt.Printf("checking for presence of `pkg` tool ... ")
	_, e := kos.Blocking(userID, "pkg", "pkg -v")
	if e != nil {
		log.Fatal(errors.Wrap(e, "ensure that the `pkg` tool is installed correctly. You can get it from here https://github.com/zeit/pkg or by running `npm install -g pkg`"))
	}
	fmt.Printf("done\n")
}

func downloadCcxtSource(kos *kelpos.KelpOS, userID string, downloadDir string) {
	fmt.Printf("making directory where we can download ccxt file %s ... ", downloadDir)
	_, e := kos.Blocking(userID, "mkdir", fmt.Sprintf("mkdir -p %s", downloadDir))
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not make directory for downloadDir "+downloadDir))
	}
	fmt.Printf("done\n")

	fmt.Printf("downloading file from URL %s ... ", ccxtDownloadURL)
	downloadFilePath := filepath.Join(downloadDir, ccxtDownloadFilename)
	e = networking.DownloadFile(ccxtDownloadURL, downloadFilePath)
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not download ccxt tar.gz file"))
	}
	fmt.Printf("done\n")

	fmt.Printf("untaring file %s ... ", downloadFilePath)
	_, e = kos.Blocking(userID, "tar", fmt.Sprintf("tar xvf %s -C %s", downloadFilePath, downloadDir))
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not untar ccxt file"))
	}
	fmt.Printf("done\n")
}

func npmInstall(kos *kelpos.KelpOS, userID string, installDir string) {
	fmt.Printf("running npm install on directory %s ... ", installDir)
	npmCmd := fmt.Sprintf("cd %s && npm install && cd -", installDir)
	_, e := kos.Blocking(userID, "npm", npmCmd)
	if e != nil {
		log.Fatal(errors.Wrap(e, "failed to run npm install"))
	}
	fmt.Printf("done\n")
}

// pkg --targets node8-linux-x64 build/ccxt/ccxt-rest-0.0.4
func runPkgTool(kos *kelpos.KelpOS, userID string, sourceDir string, outDir string, pkgos string) {
	target := fmt.Sprintf("node8-%s-x64", pkgos)

	fmt.Printf("running pkg tool on source directory %s with output directory as %s on target platform %s ... ", sourceDir, outDir, target)
	pkgCommand := fmt.Sprintf("pkg --out-path %s --targets %s %s", outDir, target, sourceDir)
	outputBytes, e := kos.Blocking(userID, "pkg", pkgCommand)
	if e != nil {
		log.Fatal(errors.Wrap(e, "failed to run pkg tool"))
	}
	fmt.Printf("done\n\n")

	pkgCmdOutput := string(outputBytes)
	log.Printf("output of pkg script:\n%s", pkgCmdOutput)

	copyDependencyFiles(kos, userID, outDir, pkgCmdOutput)
}

func copyDependencyFiles(kos *kelpos.KelpOS, userID string, outDir string, pkgCmdOutput string) {
	fmt.Println()
	fmt.Printf("copying dependency files to the output directory %s ...\n", outDir)
	for _, line := range strings.Split(pkgCmdOutput, "\n") {
		if !strings.Contains(line, "node_modules") {
			continue
		}
		filename := strings.TrimSpace(strings.Replace(line, "(MISSING)", "", -1))
		filename = strings.TrimSpace(strings.Replace(filename, "%1:", "", -1))

		fmt.Printf("    copying file %s to the output directory %s ... ", filename, outDir)
		cpCmd := fmt.Sprintf("cp %s %s", filename, outDir)
		_, e := kos.Blocking(userID, "cp", cpCmd)
		if e != nil {
			log.Fatal(errors.Wrap(e, "failed to copy dependency file "+filename))
		}
		fmt.Printf("done\n")
	}
	fmt.Printf("done\n")
	fmt.Println()
}

func mkDir(kos *kelpos.KelpOS, userID string, zipDir string) {
	fmt.Printf("making directory %s ... ", zipDir)
	_, e := kos.Blocking(userID, "mkdir", fmt.Sprintf("mkdir -p %s", zipDir))
	if e != nil {
		log.Fatal(errors.Wrap(e, "unable to make directory "+zipDir))
	}
	fmt.Printf("done\n")
}

func zipOutput(kos *kelpos.KelpOS, userID string, ccxtDir string, sourceDir string, zipFoldername string, zipOutDir string) {
	zipFilename := zipFoldername + ".zip"
	fmt.Printf("zipping directory %s as file %s ... ", filepath.Join(ccxtDir, ccxtBinOutputDir), zipFilename)
	zipCmd := fmt.Sprintf("cd %s && mv %s %s && zip -rq %s %s && cd - && mv %s %s", ccxtDir, ccxtBinOutputDir, zipFoldername, zipFilename, zipFoldername, filepath.Join(ccxtDir, zipFilename), zipOutDir)
	_, e := kos.Blocking(userID, "zip", zipCmd)
	if e != nil {
		log.Fatal(errors.Wrap(e, "unable to zip folder with ccxt binary and dependencies"))
	}
	fmt.Printf("done\n")

	zipDirPath := filepath.Join(ccxtDir, zipFoldername)
	fmt.Printf("clean up zipped directory %s ... ", zipDirPath)
	cleanupCmd := fmt.Sprintf("rm %s/* && rmdir %s", zipDirPath, zipDirPath)
	_, e = kos.Blocking(userID, "zip", cleanupCmd)
	if e != nil {
		log.Fatal(errors.Wrap(e, fmt.Sprintf("unable to cleanup zip folder %s with ccxt binary and dependencies", zipDirPath)))
	}
	fmt.Printf("done\n")
}

func generateCcxtBinary(kos *kelpos.KelpOS, userID string, pkgos string, zipFoldername string) {
	checkNodeVersion(kos, userID)
	checkPkgTool(kos, userID)

	ccxtDir := filepath.Join(kelpPrefsDirectory, "ccxt")
	sourceDir := filepath.Join(ccxtDir, ccxtUntaredDirName)
	outDir := filepath.Join(ccxtDir, ccxtBinOutputDir)
	zipOutDir := filepath.Join(ccxtDir, "zipped")

	downloadCcxtSource(kos, userID, ccxtDir)
	npmInstall(kos, userID, sourceDir)
	runPkgTool(kos, userID, sourceDir, outDir, pkgos)
	mkDir(kos, userID, zipOutDir)
	zipOutput(kos, userID, ccxtDir, outDir, zipFoldername, zipOutDir)
}
