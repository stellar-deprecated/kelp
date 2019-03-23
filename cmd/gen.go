package cmd

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generates HTML/JS/CSS files for Kelp GUI",
}

func init() {
	genCmd.Run = func(ccmd *cobra.Command, args []string) {
		fs := http.Dir("./gui/build")
		e := vfsgen.Generate(fs, vfsgen.Options{})
		if e != nil {
			log.Fatalln(e)
		}
	}
}
