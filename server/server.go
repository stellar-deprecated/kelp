package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/interstellar/kelp/support/utils"
	toml "github.com/pelletier/go-toml"
	"github.com/r3labs/sse"
	"github.com/rs/cors"
	"github.com/stellar/go/clients/horizon"
	"log"
	"net/http"
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

func getURLParam(r *http.Request, key string) string {
	keys, ok := r.URL.Query()[key]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param " + key + " is missing")
		return ""
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	return keys[0]
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
	projectId := getURLParam(r, "project")

	// don't hang here, we don't need a result
	// also elliminates zombies as it calls .Wait()
	go runTool("kelp", "trade", "--botConf", configPath("botConf", projectId), "--strategy", tradeType, "--stratConf", configPath(tradeType, projectId))

	delayedSendEvent()

	w.Write([]byte(tradeType + " started"))
}

func launchDelete(w http.ResponseWriter, r *http.Request) {
	projectId := getURLParam(r, "project")

	// don't hang here, we don't need a result
	// also elliminates zombies as it calls .Wait()
	go runTool("kelp", "trade", "--botConf", configPath("botConf", projectId), "--strategy", "delete")

	delayedSendEvent()

	w.Write([]byte("trade deleted"))
}

func getOffers(w http.ResponseWriter, r *http.Request) {
	projectId := getURLParam(r, "project")

	t, err := toml.TreeFromMap(configFields(projectId))
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
