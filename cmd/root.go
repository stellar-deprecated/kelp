package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stellar/kelp/gui/backend"
	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

// build flags
var version string
var guiVersion string
var gitBranch string
var gitHash string
var buildDate string
var env string
var amplitudeAPIKey string
var goarm string
var buildType string // set from the build script, cli or gui

const envRelease = "release"
const envDev = "dev"
const rootShort = "Kelp is a free and open-source trading bot for the Stellar universal marketplace."
const rootLong = `Kelp is a free and open-source trading bot for the Stellar universal marketplace (https://stellar.org).

Learn more about Stellar : https://www.stellar.org
Learn more about Kelp    : https://github.com/stellar/kelp`
const kelpExamples = tradeExamples + "\n  kelp trade --help"

// RootCmd is the main command for this repo
var RootCmd = &cobra.Command{
	Use:     "kelp",
	Short:   rootShort,
	Long:    rootLong,
	Example: kelpExamples,
	Run: func(ccmd *cobra.Command, args []string) {
		intro := `
  __        _______ _     ____ ___  __  __ _____    _____ ___      _  _______ _     ____  
  \ \      / / ____| |   / ___/ _ \|  \/  | ____|  |_   _/ _ \    | |/ / ____| |   |  _ \ 
   \ \ /\ / /|  _| | |  | |  | | | | |\/| |  _|      | || | | |   | ' /|  _| | |   | |_) |
    \ V  V / | |___| |__| |__| |_| | |  | | |___     | || |_| |   | . \| |___| |___|  __/ 
     \_/\_/  |_____|_____\____\___/|_|  |_|_____|    |_| \___/    |_|\_\_____|_____|_|    
																			cli=` + version + `
																			gui=` + guiVersion + `
`
		fmt.Println(intro)

		if buildType == "gui" {
			// if this is the GUI binary then we want to start off with the server command
			serverCmd.Run(ccmd, args)
		} else if buildType == "cli" {
			// else start off with the help command
			e := ccmd.Help()
			if e != nil {
				panic(e)
			}
		} else {
			panic(fmt.Sprintf("unrecognized buildType: %s", buildType))
		}
	},
}

var rootCcxtRestURL *string

func init() {
	validateBuild()
	backend.SetVersionString(guiVersion, version)

	rootCcxtRestURL = RootCmd.PersistentFlags().String("ccxt-rest-url", "", "URL to use for the CCXT-rest API. Takes precendence over the CCXT_REST_URL param set in the botConfg file for the trade command and passed as a parameter into the Kelp subprocesses started by the GUI (default URL is https://localhost:3000)")

	RootCmd.AddCommand(tradeCmd)
	RootCmd.AddCommand(serverCmd)
	RootCmd.AddCommand(strategiesCmd)
	RootCmd.AddCommand(exchangesCmd)
	RootCmd.AddCommand(terminateCmd)
	RootCmd.AddCommand(versionCmd)
}

func checkInitRootFlags() {
	if *rootCcxtRestURL != "" {
		*rootCcxtRestURL = strings.TrimSuffix(*rootCcxtRestURL, "/")
		if !strings.HasPrefix(*rootCcxtRestURL, "http://") && !strings.HasPrefix(*rootCcxtRestURL, "https://") {
			log.Printf("'ccxt-rest-url' argument must start with either `http://` or `https://`")
			panic("'ccxt-rest-url' argument must start with either `http://` or `https://`")
		}

		e := isCcxtUp(*rootCcxtRestURL)
		if e != nil {
			log.Printf(e.Error())
			panic(e)
		}

		e = sdk.SetBaseURL(*rootCcxtRestURL)
		if e != nil {
			log.Printf("unable to set CCXT-rest URL to '%s': %s", *rootCcxtRestURL, e)
			panic(fmt.Errorf("unable to set CCXT-rest URL to '%s': %s", *rootCcxtRestURL, e))
		}
	}
	// do not set rootCcxtRestURL if not specified in config so each command can handle defaults accordingly
}

func validateBuild() {
	if version == "" || guiVersion == "" || buildDate == "" || gitBranch == "" || gitHash == "" {
		fmt.Println("version information not included, please build using the build script (scripts/build.sh)")
		os.Exit(1)
	}

	if amplitudeAPIKey == "" && env == envRelease {
		utils.PrintErrorHintf("amplitude API key not included, please export AMPLITUDE_API_KEY before running build script (scripts/build.sh)")
		os.Exit(1)
	}
}

func isCcxtUp(ccxtURL string) error {
	e := networking.JSONRequest(http.DefaultClient, "GET", ccxtURL, "", map[string]string{}, nil, "")
	if e != nil {
		return fmt.Errorf("unable to connect to ccxt at the URL %s: %s", ccxtURL, e)
	}
	return nil
}
