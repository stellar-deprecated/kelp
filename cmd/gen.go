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
}
