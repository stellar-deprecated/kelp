package server

import (
	"fmt"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var CONFIG_DIR_NAME = ".kelp"
var cachedConfigDir string

func getProjectFromCmd(cmd []string) string {
	found := false
	configsDir := configDirectory()

	for i := range cmd {
		item := cmd[i]

		if found == true {
			if strings.HasPrefix(item, configsDir) {

				// strip off prefix, get next path segment
				item = strings.TrimPrefix(item, configsDir)
				item = strings.TrimPrefix(item, "/")

				// trim off after /
				split := strings.SplitN(item, "/", 2)

				if len(split) > 0 {
					return split[0]
				}

				return item
			}
		} else if item == "--botConf" {
			found = true
		}
	}

	return ""
}

func configDirectory() string {
	if len(cachedConfigDir) == 0 {
		usr, _ := user.Current()
		cachedConfigDir = filepath.Join(usr.HomeDir, CONFIG_DIR_NAME)

		// on docker the configs are located at /
		if _, err := os.Stat(cachedConfigDir); os.IsNotExist(err) {
			cachedConfigDir = "/" + CONFIG_DIR_NAME
		}
	}

	return cachedConfigDir
}

func configPath(id string, projectId string) string {
	result := ""

	configsDir := configDirectory()

	// get project folder, will be a param, for now just use default
	dirName := projectId
	if len(dirName) == 0 {
		dirName = "default"
	}

	configsDir = filepath.Join(configsDir, dirName)

	switch id {
	case "botConf":
		result = filepath.Join(configsDir, "trader.toml")
		break
	case "sell":
		result = filepath.Join(configsDir, "sell.toml")
		break
	case "mirror":
		result = filepath.Join(configsDir, "mirror.toml")
		break
	case "balanced":
		result = filepath.Join(configsDir, "balanced.toml")
		break
	case "buysell":
		result = filepath.Join(configsDir, "buysell.toml")
		break
	default:
		break
	}

	return result
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	projectId := getURLParam(r, "project")

	t, err := toml.TreeFromMap(configFields(projectId))
	if err != nil {
		log.Println(fmt.Errorf("error config file: %s \n", err))
	}

	// log.Println(t.Get("horizon_url"))

	w.Write([]byte(t.String()))
}

func configFields(projectId string) map[string]interface{} {
	configPath := configPath("botConf", projectId)

	nameNoExt := filepath.Base(configPath)
	nameNoExt = strings.TrimSuffix(nameNoExt, filepath.Ext(configPath))

	// will fail if you just use the singleton 'viper.', it won't bother reading a different config
	freshViper := viper.New()

	freshViper.SetConfigName(nameNoExt)
	freshViper.AddConfigPath(filepath.Dir(configPath))
	err := freshViper.ReadInConfig()
	if err != nil {
		log.Println(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return freshViper.AllSettings()
}
