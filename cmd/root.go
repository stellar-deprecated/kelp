package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// build flags
var version string
var buildDate string
var gitHash string

const rootShort = "Kelp is a free and open-source trading bot for the Stellar universal marketplace."
const rootLong = `Kelp is a free and open-source trading bot for the Stellar universal marketplace (https://stellar.org).

Learn more about Stellar : https://www.stellar.org
Learn more about Kelp    : https://github.com/lightyeario/kelp`
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

`
		fmt.Println(intro)
		e := ccmd.Help()
		if e != nil {
			log.Fatal(e)
		}

		fmt.Println("version:", version)
		fmt.Println("build date:", buildDate)
		fmt.Println("git hash:", gitHash)
	},
}

func init() {
	RootCmd.AddCommand(tradeCmd)
	RootCmd.AddCommand(strategiesCmd)
	RootCmd.AddCommand(exchanagesCmd)
	RootCmd.AddCommand(terminateCmd)
}
