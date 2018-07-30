package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

const rootShort = "Kelp is a free and open-source trading bot for the Stellar universal marketplace."
const rootLong = `Kelp is a free and open-source trading bot for the Stellar universal marketplace.
Learn more about Kelp here: https://github.com/lightyeario/kelp`

// RootCmd is the main command for this repo
var RootCmd = &cobra.Command{
	Use:   "kelp",
	Short: rootShort,
	Long:  rootLong,
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
	},
}

func init() {
	RootCmd.AddCommand(tradeCmd)
	RootCmd.AddCommand(terminateCmd)
}
