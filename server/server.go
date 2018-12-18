package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/interstellar/kelp/support/utils"
	toml "github.com/pelletier/go-toml"
	"github.com/r3labs/sse"
	"github.com/rs/cors"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/stellar/go/clients/horizon"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// global vars
var sseServer *sse.Server

func Start() {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	})
	r.Use(c.Handler)

	r.Get("/", getHome)
	r.Get("/help", getHelp)
	r.Get("/version", getVersion)
	r.Get("/strategies", getStrategies)
	r.Get("/exchanges", getExchanges)
	r.Get("/buysell", launchBuySell)
	r.Get("/sell", launchSell)
	r.Get("/mirror", launchMirror)
	r.Get("/balanced", launchBalanced)
	r.Get("/delete", launchDelete)
	r.Get("/list", getProcesses)
	r.Get("/offers", getOffers)
	r.Put("/params", launchWithParams)
	r.Get("/config", getConfig)
	r.Put("/kill", killKelp)

	// sse, use http://server/events?stream=messages
	sseServer = sse.New()
	sseServer.AutoReplay = false // must turn this off or all the pings will replay for every client refresh
	sseServer.CreateStream("messages")
	r.Get("/events", sseServer.HTTPHandler)

	http.ListenAndServe(":8991", r)
}

func delayedSendEvent() {
	time.AfterFunc(1*time.Second, sendEvent)
}

func sendEvent() {
	sseServer.Publish("messages", &sse.Event{
		Data: []byte("ping"),
	})
}

func launchWithParams(w http.ResponseWriter, r *http.Request) {
	// result := chi.URLParam(r, "kelp")
	// requestDump, _ := httputil.DumpRequest(r, true)
	type Message struct {
		Kelp string
	}
	var m Message
	json.NewDecoder(r.Body).Decode(&m)

	stringSlice := strings.Split(m.Kelp, " ")

	result := runKelp(stringSlice...)

	w.Write([]byte(result))
}

func killKelp(w http.ResponseWriter, r *http.Request) {
	type Message struct {
		Pid string // pid of kelp to kill
	}
	var m Message
	json.NewDecoder(r.Body).Decode(&m)

	if len(m.Pid) > 0 {
		runTool("kill", m.Pid) // -15 SIGTERM default
	} else {
		log.Println("kill pid was invalid")
	}

	delayedSendEvent()

	w.Write([]byte("killed: " + m.Pid))
}

func getHome(w http.ResponseWriter, r *http.Request) {
	result := runKelp("")

	w.Write([]byte(result))
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	result := runKelp("version")

	w.Write([]byte(result))
}

func getHelp(w http.ResponseWriter, r *http.Request) {
	result := runKelp("help", "trade")

	w.Write([]byte(result))
}

func getStrategies(w http.ResponseWriter, r *http.Request) {
	result := runKelp("strategies")

	w.Write([]byte(result))
}

func getExchanges(w http.ResponseWriter, r *http.Request) {
	result := runKelp("exchanges")

	w.Write([]byte(result))
}

func configPath(id string) string {
	result := ""

	usr, _ := user.Current()
	configsDir := filepath.Join(usr.HomeDir, ".kelp")

	// on docker the configs are located at /configs, otherwise ./configs
	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		configsDir = "/configs"
	}

	// get project folder, will be a param, for now just use default
	configsDir = filepath.Join(configsDir, "default")

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

func launchBuySell(w http.ResponseWriter, r *http.Request) {
	launchTrade(w, r, "buysell")
}

func launchSell(w http.ResponseWriter, r *http.Request) {
	launchTrade(w, r, "sell")
}

func launchMirror(w http.ResponseWriter, r *http.Request) {
	launchTrade(w, r, "mirror")
}

func launchBalanced(w http.ResponseWriter, r *http.Request) {
	launchTrade(w, r, "balanced")
}

func launchTrade(w http.ResponseWriter, r *http.Request, tradeType string) {
	// don't hang here, we don't need a result
	// also elliminates zombies as it calls .Wait()
	go runTool("kelp", "trade", "--botConf", configPath("botConf"), "--strategy", tradeType, "--stratConf", configPath(tradeType))

	delayedSendEvent()

	w.Write([]byte(tradeType + "started"))
}

func launchDelete(w http.ResponseWriter, r *http.Request) {
	// don't hang here, we don't need a result
	// also elliminates zombies as it calls .Wait()
	go runTool("kelp", "trade", "--botConf", configPath("botConf"), "--strategy", "delete")

	delayedSendEvent()

	w.Write([]byte("trade deleted"))
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	t, err := toml.TreeFromMap(configFields())
	if err != nil {
		log.Println(fmt.Errorf("error config file: %s \n", err))
	}

	// log.Println(t.Get("horizon_url"))

	w.Write([]byte(t.String()))
}

func configFields() map[string]interface{} {
	configPath := configPath("botConf")

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

func getOffers(w http.ResponseWriter, r *http.Request) {
	t, err := toml.TreeFromMap(configFields())
	if err != nil {
		log.Println(fmt.Errorf("error config file: %s \n", err))
	}

	horizonURL := t.Get("horizon_url").(string)
	seed := t.Get("trading_secret_seed").(string)

	client := &horizon.Client{
		URL:  horizonURL,
		HTTP: http.DefaultClient,
	}

	sourceAccount, _ := utils.ParseSecret(seed)

	offers, _ := utils.LoadAllOffers(*sourceAccount, client)
	js, _ := json.Marshal(offers)

	w.Write(js)
}

func getProcesses(w http.ResponseWriter, r *http.Request) {
	var v []*process.Process

	v, err := process.Processes()
	if err != nil {
		log.Fatal(err)
	}

	result := []map[string]string{}

	for _, p := range v {
		name, _ := p.Name()
		cmd, _ := p.Cmdline()
		cmdSlice, _ := p.CmdlineSlice()

		pid := fmt.Sprintf("%v", p.Pid)

		if name == "kelp" {
			if len(cmdSlice) > 1 && cmdSlice[1] != "serve" {
				m := make(map[string]string)
				m["pid"] = pid
				m["cmd"] = cmd
				m["name"] = name

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
