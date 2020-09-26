package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

var (
	amplitudeAPIURL string = "https://api2.amplitude.com/2/httpapi"
)

const (
	startupEventName string = "ce:test_startup"
	updateEventName  string = "ce:test_update"
	deleteEventName  string = "ce:test_delete"
)

// MetricsTracker wraps the properties for Amplitude events,
// and can be used to directly send events to the
// Amplitude HTTP API.
type MetricsTracker struct {
	client *http.Client
	apiKey string
	userID string
	props  commonProps
	start  time.Time
}

type event struct {
	UserID    string      `json:"user_id"`
	DeviceID  string      `json:"device_id"`
	EventType string      `json:"event_type"`
	Props     interface{} `json:"event_properties"`
}

// props holds the properties that we need for all Amplitude events.
// This lives on the `MetricsTracker` struct.
// TODO: Add geodata.
// TODO: Add cloud server information.
type commonProps struct {
	CLIVersion         string    `json:"cli_version"`
	Goos               string    `json:"goos"`
	Goarch             string    `json:"goarch"`
	Goarm              string    `json:"goarm"`
	GuiVersion         string    `json:"gui_version"`
	Strategy           string    `json:"strategy"`
	UpdateTimeInterval int32     `json:"update_time_interval"`
	Exchange           string    `json:"exchange"`
	TradingPair        string    `json:"trading_pair"`
	SessionID          uuid.UUID `json:"session_id"`
	SecondsSinceStart  float64   `json:"seconds_since_start"`
}

// deleteProps holds the properties for the delete Amplitude event.
// TODO: StackTrace may need to be a message instead of or in addition to a
// stack trace. The goal is to get crash logs, Amplitude may not enable this.
type deleteProps struct {
	commonProps
	Exit       bool   `json:"exit"`
	StackTrace string `json:"stack_trace"`
}

// MakeMetricsTracker is a factory method to create a `metrics.Tracker`.
func MakeMetricsTracker(
	userID string,
	apiKey string,
	client *http.Client,
	start time.Time,
	version string,
	goos string,
	goarch string,
	goarm string,
	guiVersion string,
	strategy string,
	updateTimeInterval int32,
	exchange string,
	tradingPair string,
) (*MetricsTracker, error) {
	sessionID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("could not generate uuid with error %s", err)
	}
	props := commonProps{
		CLIVersion:         version,
		Goos:               goos,
		Goarch:             goarch,
		Goarm:              goarm,
		GuiVersion:         guiVersion,
		Strategy:           strategy,
		UpdateTimeInterval: updateTimeInterval,
		Exchange:           exchange,
		TradingPair:        tradingPair,
		SessionID:          sessionID,
	}

	return &MetricsTracker{
		client: client,
		apiKey: apiKey,
		props:  props,
		start:  start,
	}, nil
}

// SendStartupEvent sends the startup Amplitude event.
func (mt *MetricsTracker) SendStartupEvent() error {
	return mt.sendEvent(startupEventName, mt.props)
}

// SendUpdateEvent sends the update Amplitude event.
func (mt *MetricsTracker) SendUpdateEvent(t time.Time) error {
	updateProps := mt.props
	updateProps.SecondsSinceStart = t.Sub(mt.start).Seconds()
	return mt.sendEvent(updateEventName, updateProps)
}

// SendDeleteEvent sends the delete Amplitude event.
func (mt *MetricsTracker) SendDeleteEvent(exit bool) error {
	commonProps := mt.props
	commonProps.SecondsSinceStart = time.Now().Sub(mt.start).Seconds()
	deleteProps := deleteProps{
		commonProps: commonProps,
		Exit:        exit,
		StackTrace:  string(debug.Stack()),
	}

	return mt.sendEvent(deleteEventName, deleteProps)
}

// TODO: Re-implement using `networking/JSONRequestDynamicHeaders`.
func (mt *MetricsTracker) sendEvent(eventType string, eventProps interface{}) error {
	requestBody, e := json.Marshal(map[string]interface{}{
		"api_key": mt.apiKey,
		"events": []event{event{
			UserID:    "12345", // TODO: Determine actual user id.
			EventType: eventType,
			Props:     eventProps,
		}},
	})

	if e != nil {
		log.Print("could not marshal json request")
		return fmt.Errorf("could not marshal json request: %s", e)
	}

	resp, e := http.Post(amplitudeAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if e != nil {
		log.Print("could not post amplitude request")
		return fmt.Errorf("could not post amplitude request: %s", e)
	}

	defer resp.Body.Close()

	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		log.Print("could not read response body")
		return fmt.Errorf("could not read response body: %s", e)
	}

	log.Printf("Successfully sent event metric of type %s", eventType)
	return nil
}
