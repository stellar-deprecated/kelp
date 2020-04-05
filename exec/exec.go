package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/stellar/kelp/support/kelpos"
)

func main() {
	flag.Parse()
	cmd := ""
	for _, arg := range flag.Args() {
		cmd += fmt.Sprintf(" %s", arg)
	}
	if cmd == "" {
		log.Fatal("there were no arguments, needs at least one command line argument to run as a command on the OS")
	}

	kos := kelpos.GetKelpOS()
	outputBytes, e := kos.Blocking("cmd", cmd)
	if e != nil {
		log.Fatal(e)
	}
	outputString := string(outputBytes)
	fmt.Println(outputString)
}
