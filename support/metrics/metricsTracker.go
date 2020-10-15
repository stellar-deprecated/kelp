package metrics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/stellar/kelp/support/networking"
)

// Custom events in Amplitude should be named with "ce:event_name",
// so the web console displays it as "[Custom] event_name".
const (
	amplitudeAPIURL  string = "https://api2.amplitude.com/2/httpapi"
	startupEventName string = "ce:test_startup"
	updateEventName  string = "ce:test_update"
	deleteEventName  string = "ce:test_delete"
)

// MetricsTracker wraps the properties for Amplitude events,
// and can be used to directly send events to the
// Amplitude HTTP API.
type MetricsTracker struct {
	client     *http.Client
	apiKey     string
	userID     string
	deviceID   string
	props      commonProps
	start      time.Time
	isDisabled bool
}

// TODO DS Investigate other fields to add to this top-level event.
// fields for the event object: https://help.amplitude.com/hc/en-us/articles/360032842391-HTTP-API-V2#http-api-v2-events
type event struct {
	UserID    string      `json:"user_id"`
	SessionID int64       `json:"session_id"`
	DeviceID  string      `json:"device_id"`
	EventType string      `json:"event_type"`
	Version   string      `json:"app_version"`
	Props     interface{} `json:"event_properties"`
}

// props holds the properties that we need for all Amplitude events.
// This lives on the `MetricsTracker` struct.
// TODO DS Add geodata.
// TODO DS Add cloud server information.
// TODO DS Add time to run update function as `millisForUpdate`.
type commonProps struct {
	CliVersion                string  `json:"cli_version"`
	Goos                      string  `json:"goos"`
	Goarch                    string  `json:"goarch"`
	Goarm                     string  `json:"goarm"`
	GuiVersion                string  `json:"gui_version"`
	Strategy                  string  `json:"strategy"`
	UpdateTimeIntervalSeconds int32   `json:"update_time_interval_seconds"`
	Exchange                  string  `json:"exchange"`
	TradingPair               string  `json:"trading_pair"`
	SecondsSinceStart         float64 `json:"seconds_since_start"`
	IsTestnet                 bool    `json:"is_testnet"`
}

// updateProps holds the properties for the update Amplitude event.
type updateProps struct {
	commonProps
	Success         bool  `json:"success"`
	MillisForUpdate int64 `json:"millis_for_update"`
}

// deleteProps holds the properties for the delete Amplitude event.
// TODO DS StackTrace may need to be a message instead of or in addition to a
// stack trace. The goal is to get crash logs, Amplitude may not enable this.
type deleteProps struct {
	commonProps
	Exit       bool   `json:"exit"`
	StackTrace string `json:"stack_trace"`
}

type eventWrapper struct {
	ApiKey string  `json:"api_key"`
	Events []event `json:"events"`
}

// response structure taken from here: https://help.amplitude.com/hc/en-us/articles/360032842391-HTTP-API-V2#tocSsuccesssummary
type amplitudeResponse struct {
	Code             int   `json:"code"`
	EventsIngested   int   `json:"events_ingested"`
	PayloadSizeBytes int   `json:"payload_size_bytes"`
	ServerUploadTime int64 `json:"server_upload_time"`
}

// String is the Stringer method
func (ar amplitudeResponse) String() string {
	return fmt.Sprintf("amplitudeResponse[Code=%d, EventsIngested=%d, PayloadSizeBytes=%d, ServerUploadTime=%d (%s)]",
		ar.Code,
		ar.EventsIngested,
		ar.PayloadSizeBytes,
		ar.ServerUploadTime,
		time.Unix(ar.ServerUploadTime, 0).Format("20060102T150405MST"),
	)
}

// MakeMetricsTracker is a factory method to create a `metrics.Tracker`.
func MakeMetricsTracker(
	userID string,
	deviceID string,
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
	isDisabled bool,
	isTestnet bool,
) (*MetricsTracker, error) {
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
		IsTestnet:                 isTestnet,
	}

	return &MetricsTracker{
		client:     client,
		apiKey:     apiKey,
		userID:     userID,
		deviceID:   deviceID,
		props:      props,
		start:      start,
		isDisabled: isDisabled,
	}, nil
}

// SendStartupEvent sends the startup Amplitude event.
func (mt *MetricsTracker) SendStartupEvent() error {
	return mt.sendEvent(startupEventName, mt.props)
}

// SendUpdateEvent sends the update Amplitude event.
func (mt *MetricsTracker) SendUpdateEvent(now time.Time, success bool, millisForUpdate int64) error {
	commonProps := mt.props
	commonProps.SecondsSinceStart = now.Sub(mt.start).Seconds()
	updateProps := updateProps{
		commonProps:     commonProps,
		Success:         success,
		MillisForUpdate: millisForUpdate,
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

func (mt *MetricsTracker) sendEvent(eventType string, eventProps interface{}) error {
	if mt.apiKey == "" || mt.userID == "-1" || mt.isDisabled {
		log.Printf("metric - not sending event metric of type '%s' because metrics are disabled", eventType)
		return nil
	}

	// session_id is the start time of the session in milliseconds since epoch (Unix Timestamp),
	// necessary to associate events with a particular system (taken from amplitude docs)
	eventW := eventWrapper{
		ApiKey: mt.apiKey,
		Events: []event{{
			UserID:    mt.userID,
			SessionID: mt.start.Unix() * 1000, // convert to millis based on docs
			DeviceID:  mt.deviceID,
			EventType: eventType,
			Props:     eventProps,
			Version:   mt.props.CliVersion,
		}},
	}
	requestBody, e := json.Marshal(eventW)
	if e != nil {
		return fmt.Errorf("could not marshal json request: %s", e)
	}

	// TODO DS - wrap these API functions into support/sdk/amplitude.go
	var responseData amplitudeResponse
	e = networking.JSONRequest(mt.client, "POST", amplitudeAPIURL, string(requestBody), map[string]string{}, &responseData, "")
	if e != nil {
		return fmt.Errorf("could not post amplitude request: %s", e)
	}

	if responseData.Code == 200 {
		log.Printf("metric - successfully sent event metric of type '%s'", eventType)
	} else {
		// work on copy so we don't modify original (good hygiene)
		eventWCensored := *(&eventW)
		// we don't want to display the apiKey in the logs so censor it
		eventWCensored.ApiKey = ""
		requestWCensored, e := json.Marshal(eventWCensored)
		if e != nil {
			log.Printf("metric - failed to send event metric of type '%s' (response=%s), error while trying to marshall requestWCensored: %s", eventType, responseData.String(), e)
		} else {
			log.Printf("metric - failed to send event metric of type '%s' (requestWCensored=%s; response=%s)", eventType, string(requestWCensored), responseData.String())
		}
	}
	return nil
}
