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

var environmentsDarwin = `{
	"environments": [
		{"os": "darwin", "arch": "amd64"}
	]
}`

var environmentsLinux = `{
	"environments": [
		{"os": "linux", "arch": "amd64"}
	]
}`

var environmentsWindows = `{
	"environments": [
		{"os": "windows", "arch": "amd64"}
	]
}`

func main() {
	buildAllP := flag.Bool("a", false, "whether to build for all platforms (default builds only for native platform)")
	buildPlatformP := flag.String("p", "", "explicitly specify a specific platform to build for")
	flag.Parse()
	buildAll := *buildAllP

	var bundlerJSON map[string]interface{}
	e := json.Unmarshal([]byte(bundler), &bundlerJSON)
	if e != nil {
		panic(e)
	}

	if *buildPlatformP == "darwin" {
		setPlatform(environmentsDarwin, bundlerJSON)
	} else if *buildPlatformP == "linux" {
		setPlatform(environmentsLinux, bundlerJSON)
	} else if *buildPlatformP == "windows" {
		setPlatform(environmentsWindows, bundlerJSON)
	} else if buildAll {
		setPlatform(environments, bundlerJSON)
	} // else only for native platform

	jsonBytes, e := json.MarshalIndent(bundlerJSON, "", "    ")
	if e != nil {
		panic(e)
	}
	jsonString := string(jsonBytes)
	fmt.Println(jsonString)
}

func setPlatform(envs string, bundlerJSON map[string]interface{}) {
	var environmentsJSON map[string]interface{}
	e := json.Unmarshal([]byte(envs), &environmentsJSON)
	if e != nil {
		panic(e)
	}

	for k, v := range environmentsJSON {
		bundlerJSON[k] = v
	}
}
