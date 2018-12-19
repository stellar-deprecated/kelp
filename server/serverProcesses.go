package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"log"
	"net/http"
	"os/exec"
)

func getProcesses(w http.ResponseWriter, r *http.Request) {
	var v []*process.Process

	v, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}

	result := []map[string]string{}

	for _, p := range v {
		name, _ := p.Name()

		if name == "kelp" {
			cmdSlice, _ := p.CmdlineSlice()

			if len(cmdSlice) > 1 && cmdSlice[1] != "serve" {
				pid := fmt.Sprintf("%v", p.Pid)
				cmd, _ := p.Cmdline()
				project := getProjectFromCmd(cmdSlice)

				m := make(map[string]string)
				m["pid"] = pid
				m["cmd"] = cmd
				m["name"] = name
				m["project"] = project

				result = append(result, m)
			}
		}
	}

	js, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(js)
}

func runKelp(params ...string) string {
	return runTool("kelp", params...)
}

func runTool(tool string, params ...string) string {
	debug := false
	if debug {
		log.Println(tool)
		for _, v := range params {
			log.Println(v)
		}
	}

	cmd := exec.Command(tool, params...)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		log.Println(stdErr.String())

		// kill returns an err?  Don't put fatal here unless you test killKelp
		log.Println(err)
	}

	return stdOut.String()
}
