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

func configPath(id string, projectID string) string {
	result := ""

	usr, _ := user.Current()
	configsDir := filepath.Join(usr.HomeDir, ".kelp")

	// on docker the configs are located at /configs, otherwise ./configs
	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		configsDir = "/configs"
	}

	// get project folder, will be a param, for now just use default
	dirName := projectID
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

	viper.SetConfigName(nameNoExt)
	viper.AddConfigPath(filepath.Dir(configPath))
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	return viper.AllSettings()
}
