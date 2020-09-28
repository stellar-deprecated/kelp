package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/stellar/kelp/support/networking"

	"github.com/google/uuid"
)

var (
	amplitudeAPIURL string = "https://api2.amplitude.com/2/httpapi"
)

// Custom events in Amplitude should be named with "ce:event_name",
// so the web console displays it as "[Custom] event_name".
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
// TODO DS Add geodata.
// TODO DS Add cloud server information.
type commonProps struct {
	CliVersion                string    `json:"cli_version"`
	Goos                      string    `json:"goos"`
	Goarch                    string    `json:"goarch"`
	Goarm                     string    `json:"goarm"`
	GuiVersion                string    `json:"gui_version"`
	Strategy                  string    `json:"strategy"`
	UpdateTimeIntervalSeconds int32     `json:"update_time_interval_seconds"`
	Exchange                  string    `json:"exchange"`
	TradingPair               string    `json:"trading_pair"`
	SessionID                 uuid.UUID `json:"session_id"`
	SecondsSinceStart         float64   `json:"seconds_since_start"`
}

// updateProps holds the properties for the update Amplitude event.
type updateProps struct {
	commonProps
	Success bool `json:"success"`
}

// deleteProps holds the properties for the delete Amplitude event.
// TODO DS StackTrace may need to be a message instead of or in addition to a
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
	updateTimeIntervalSeconds int32,
	exchange string,
	tradingPair string,
) (*MetricsTracker, error) {
	sessionID, e := uuid.NewUUID()
	if e != nil {
		return nil, fmt.Errorf("could not generate uuid with error %s", e)
	}
	props := commonProps{
		CliVersion:                version,
		Goos:                      goos,
		Goarch:                    goarch,
		Goarm:                     goarm,
		GuiVersion:                guiVersion,
		Strategy:                  strategy,
		UpdateTimeIntervalSeconds: updateTimeIntervalSeconds,
		Exchange:                  exchange,
		TradingPair:               tradingPair,
		SessionID:                 sessionID,
	}

	return &MetricsTracker{
		userID: userID,
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
func (mt *MetricsTracker) SendUpdateEvent(now time.Time, success bool) error {
	commonProps := mt.props
	commonProps.SecondsSinceStart = now.Sub(mt.start).Seconds()
	updateProps := updateProps{
		commonProps: commonProps,
		Success:     success,
	}
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

// TODO DS Re-implement using `networking/JSONRequestDynamicHeaders`.
func (mt *MetricsTracker) sendEvent(eventType string, eventProps interface{}) error {
	requestBody, e := json.Marshal(map[string]interface{}{
		"api_key": mt.apiKey,
		"events": []event{event{
			UserID:    mt.userID,
			DeviceID:  mt.userID,
			EventType: eventType,
			Props:     eventProps,
		}},
	})

	if e != nil {
		log.Print("could not marshal json request")
		return fmt.Errorf("could not marshal json request: %s", e)
	}

	log.Printf("Sending request: %v\n", string(requestBody))

	var responseData interface{}
	e = networking.JSONRequest(mt.client, "POST", amplitudeAPIURL, string(requestBody), map[string]string{}, &responseData, "")
	if e != nil {
		return fmt.Errorf("could not post amplitude request: %s", e)
	}

	log.Printf("Got response: %v\n", responseData)

	log.Printf("Successfully sent event metric of type %s", eventType)
	return nil
}
