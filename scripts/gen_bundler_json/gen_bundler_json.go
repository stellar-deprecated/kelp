package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

var bundler = `{
    "app_name": "Kelp",
    "icon_path_darwin": "resources/kelp-icon@2x.icns",
    "icon_path_linux": "resources/kelp-icon@2x.png",
    "icon_path_windows": "resources/kelp-icon@2x.ico",
    "bind": {
        "output_path": "./cmd",
        "package": "cmd"
    }
}`

var environments = `{
	"environments": [
		{"os": "darwin", "arch": "amd64"},
		{"os": "linux", "arch": "amd64"},
		{"os": "windows", "arch": "amd64"}
	]
}`

func main() {
	buildAllP := flag.Bool("a", false, "whether to build for all platforms (default builds only for native platform)")
	flag.Parse()
	buildAll := *buildAllP

	var bundlerJSON map[string]interface{}
	e := json.Unmarshal([]byte(bundler), &bundlerJSON)
	if e != nil {
		panic(e)
	}

	if buildAll {
		var environmentsJSON map[string]interface{}
		e := json.Unmarshal([]byte(environments), &environmentsJSON)
		if e != nil {
			panic(e)
		}

		for k, v := range environmentsJSON {
			bundlerJSON[k] = v
		}
	}

	jsonBytes, e := json.MarshalIndent(bundlerJSON, "", "    ")
	if e != nil {
		panic(e)
	}
	jsonString := string(jsonBytes)
	fmt.Println(jsonString)
}
