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

// we don't want this to be a custom event, custom events should only be added from the amplitude UI
const (
	amplitudeAPIURL  string = "https://api2.amplitude.com/2/httpapi"
	startupEventName string = "bot_startup"
	updateEventName  string = "update_offers"
	deleteEventName  string = "delete_offers"
)

// MetricsTracker wraps the properties for Amplitude events,
// and can be used to directly send events to the
// Amplitude HTTP API.
type MetricsTracker struct {
	client              *http.Client
	apiKey              string
	userID              string
	deviceID            string
	props               commonProps
	botStartTime        time.Time
	isDisabled          bool
	updateEventSentTime *time.Time
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
	CliVersion                       string  `json:"cli_version"`
	GitHash                          string  `json:"git_hash"`
	Env                              string  `json:"env"`
	Goos                             string  `json:"goos"`
	Goarch                           string  `json:"goarch"`
	Goarm                            string  `json:"goarm"`
	GuiVersion                       string  `json:"gui_version"`
	Strategy                         string  `json:"strategy"`
	UpdateTimeIntervalSeconds        int32   `json:"update_time_interval_seconds"`
	Exchange                         string  `json:"exchange"`
	TradingPair                      string  `json:"trading_pair"`
	SecondsSinceStart                float64 `json:"seconds_since_start"`
	IsTestnet                        bool    `json:"is_testnet"`
	MaxTickDelayMillis               int64   `json:"max_tick_delay_millis"`
	SubmitMode                       string  `json:"submit_mode"`
	DeleteCyclesThreshold            int64   `json:"delete_cycles_threshold"`
	FillTrackerSleepMillis           uint32  `json:"fill_tracker_sleep_millis"`
	FillTrackerDeleteCyclesThreshold int64   `json:"fill_tracker_delete_cycles_threshold"`
	SynchronizeStateLoadEnable       bool    `json:"synchronize_state_load_enable"`
	SynchronizeStateLoadMaxRetries   int     `json:"synchronize_state_load_max_retries"`
	EnabledFeatureDollarValue        bool    `json:"enabled_feature_dollar_value"`
	AlertType                        string  `json:"alert_type"`
	EnabledFeatureMonitoring         bool    `json:"enabled_feature_monitoring"`
	EnabledFeatureFilters            bool    `json:"enabled_feature_filters"`
	EnabledFeaturePostgres           bool    `json:"enabled_feature_postgres"`
}

// updateProps holds the properties for the update Amplitude event.
type updateProps struct {
	commonProps
	Success                      bool    `json:"success"`
	MillisForUpdate              int64   `json:"millis_for_update"`
	SecondsSinceLastUpdateMetric float64 `json:"seconds_since_last_update_metric"` // helps understand total runtime of bot when summing this field across events
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
	botStartTime time.Time,
	version string,
	gitHash string,
	env string,
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
	maxTickDelayMillis int64,
	submitMode string,
	deleteCyclesThreshold int64,
	fillTrackerSleepMillis uint32,
	fillTrackerDeleteCyclesThreshold int64,
	synchronizeStateLoadEnable bool,
	synchronizeStateLoadMaxRetries int,
	enabledFeatureDollarValue bool,
	alertType string,
	enabledFeatureMonitoring bool,
	enabledFeatureFilters bool,
	enabledFeaturePostgres bool,
) (*MetricsTracker, error) {
	props := commonProps{
		CliVersion:                       version,
		GitHash:                          gitHash,
		Env:                              env,
		Goos:                             goos,
		Goarch:                           goarch,
		Goarm:                            goarm,
		GuiVersion:                       guiVersion,
		Strategy:                         strategy,
		UpdateTimeIntervalSeconds:        updateTimeIntervalSeconds,
		Exchange:                         exchange,
		TradingPair:                      tradingPair,
		IsTestnet:                        isTestnet,
		MaxTickDelayMillis:               maxTickDelayMillis,
		SubmitMode:                       submitMode,
		DeleteCyclesThreshold:            deleteCyclesThreshold,
		FillTrackerSleepMillis:           fillTrackerSleepMillis,
		FillTrackerDeleteCyclesThreshold: fillTrackerDeleteCyclesThreshold,
		SynchronizeStateLoadEnable:       synchronizeStateLoadEnable,
		SynchronizeStateLoadMaxRetries:   synchronizeStateLoadMaxRetries,
		EnabledFeatureDollarValue:        enabledFeatureDollarValue,
		AlertType:                        alertType,
		EnabledFeatureMonitoring:         enabledFeatureMonitoring,
		EnabledFeatureFilters:            enabledFeatureFilters,
		EnabledFeaturePostgres:           enabledFeaturePostgres,
	}

	return &MetricsTracker{
		client:              client,
		apiKey:              apiKey,
		userID:              userID,
		deviceID:            deviceID,
		props:               props,
		botStartTime:        botStartTime,
		isDisabled:          isDisabled,
		updateEventSentTime: nil,
	}, nil
}

// GetUpdateEventSentTime gets the last sent time of the update event.
func (mt *MetricsTracker) GetUpdateEventSentTime() *time.Time {
	return mt.updateEventSentTime
}

// SendStartupEvent sends the startup Amplitude event.
func (mt *MetricsTracker) SendStartupEvent() error {
	return mt.sendEvent(startupEventName, mt.props)
}

// SendUpdateEvent sends the update Amplitude event.
func (mt *MetricsTracker) SendUpdateEvent(now time.Time, success bool, millisForUpdate int64) error {
	commonProps := mt.props
	commonProps.SecondsSinceStart = now.Sub(mt.botStartTime).Seconds()

	var secondsSinceLastUpdateMetric float64
	if mt.updateEventSentTime == nil {
		secondsSinceLastUpdateMetric = commonProps.SecondsSinceStart
	} else {
		secondsSinceLastUpdateMetric = now.Sub(*mt.updateEventSentTime).Seconds()
	}
	updateProps := updateProps{
		commonProps:                  commonProps,
		Success:                      success,
		MillisForUpdate:              millisForUpdate,
		SecondsSinceLastUpdateMetric: secondsSinceLastUpdateMetric,
	}
	e := mt.sendEvent(updateEventName, updateProps)
	if e != nil {
		return fmt.Errorf("could not send update event: %s", e)
	}

	mt.updateEventSentTime = &now
	return nil
}

// SendDeleteEvent sends the delete Amplitude event.
func (mt *MetricsTracker) SendDeleteEvent(exit bool) error {
	commonProps := mt.props
	commonProps.SecondsSinceStart = time.Now().Sub(mt.botStartTime).Seconds()
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
			SessionID: mt.botStartTime.Unix() * 1000, // convert to millis based on docs
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
