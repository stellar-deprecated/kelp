package main

import (
	"log"

	"github.com/lightyeario/kelp/cmd"
)

func main() {
	e := cmd.RootCmd.Execute()
	if e != nil {
		log.Fatal(e)
	}
}
