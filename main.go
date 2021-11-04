package main

import (
	"log"

	"github.com/stellar/kelp/cmd"

	_ "github.com/lib/pq"
)

func main() {
	e := cmd.RootCmd.Execute()
	if e != nil {
		log.Fatal(e)
	}
}
