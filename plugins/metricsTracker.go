package plugins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/utils"
)

// we don't want this to be a custom event, custom events should only be added from the amplitude UI
const (
	amplitudeAPIURL      string = "https://api2.amplitude.com/2/httpapi"
	startupEventName     string = "bot_startup"
	updateEventName      string = "update_offers"
	deleteEventName      string = "delete_offers"
	secondsSinceStartKey string = "seconds_since_start"
	guiUserIDKey         string = "gui_user_id"
)

// MetricsTracker wraps the properties for Amplitude events,
// and can be used to directly send events to the
// Amplitude HTTP API.
type MetricsTracker struct {
	client              *http.Client
	apiKey              string
	userID              string
	guiUserID           string
	deviceID            string
	eventProps          map[string]interface{}
	botStartTime        time.Time
	isDisabled          bool
	updateEventSentTime *time.Time
	cliVersion          string
	failedStartupSend   bool
}

// TODO DS Investigate other fields to add to this top-level event.
// fields for the event object: https://help.amplitude.com/hc/en-us/articles/360032842391-HTTP-API-V2#http-api-v2-events
type event struct {
	UserID          string      `json:"user_id"`
	SessionID       int64       `json:"session_id"`
	DeviceID        string      `json:"device_id"`
	EventType       string      `json:"event_type"`
	Version         string      `json:"app_version"`
	EventProperties interface{} `json:"event_properties"`
}

// CommonPropsStruct holds the properties that are common to all our Amplitude events.
// TODO DS Add geodata.
// TODO DS Add cloud server information.
// TODO DS Add time to run update function as `millisForUpdate`.
type CommonPropsStruct struct {
	CliVersion        string  `json:"cli_version"`
	GitHash           string  `json:"git_hash"`
	Env               string  `json:"env"`
	Goos              string  `json:"goos"`
	Goarch            string  `json:"goarch"`
	Goarm             string  `json:"goarm"`
	GoVersion         string  `json:"go_version"`
	SecondsSinceStart float64 `json:"seconds_since_start"`
	IsTestnet         bool    `json:"is_testnet"`
	GuiVersion        string  `json:"gui_version"` // can be present for cli trackers if they are started from the GUI
}

// MakeCommonProps is a factory mmethod for CommonPropsStruct
func MakeCommonProps(
	cliVersion string,
	gitHash string,
	env string,
	goos string,
	goarch string,
	goarm string,
	goVersion string,
	secondsSinceStart float64,
	isTestnet bool,
	guiVersion string,
) *CommonPropsStruct {
	return &CommonPropsStruct{
		CliVersion:        cliVersion,
		GitHash:           gitHash,
		Env:               env,
		Goos:              goos,
		Goarch:            goarch,
		Goarm:             goarm,
		GoVersion:         goVersion,
		SecondsSinceStart: secondsSinceStart,
		IsTestnet:         isTestnet,
		GuiVersion:        guiVersion,
	}
}

// CliPropsStruct contains only those props that are needed for all CLI events
type CliPropsStruct struct {
	Strategy                         string  `json:"strategy"`
	UpdateTimeIntervalSeconds        float64 `json:"update_time_interval_seconds"`
	Exchange                         string  `json:"exchange"`
	TradingPair                      string  `json:"trading_pair"`
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
	EnabledFeatureLogging            bool    `json:"enabled_feature_logging"`
	OperationalBuffer                float64 `json:"operational_buffer"`
	OperationalBufferNonNativePct    float64 `json:"operational_buffer_non_native_pct"`
	SimMode                          bool    `json:"sim_mode"`
	FixedIterations                  uint64  `json:"fixed_iterations"`
}

// MakeCliProps is a factory mmethod for CommonPropsStruct
func MakeCliProps(
	strategy string,
	updateTimeIntervalSeconds float64,
	exchange string,
	tradingPair string,
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
	enabledFeatureLogging bool,
	operationalBuffer float64,
	operationalBufferNonNativePct float64,
	simMode bool,
	fixedIterations uint64,
) *CliPropsStruct {
	return &CliPropsStruct{
		Strategy:                         strategy,
		UpdateTimeIntervalSeconds:        updateTimeIntervalSeconds,
		Exchange:                         exchange,
		TradingPair:                      tradingPair,
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
		EnabledFeatureLogging:            enabledFeatureLogging,
		OperationalBuffer:                operationalBuffer,
		OperationalBufferNonNativePct:    operationalBufferNonNativePct,
		SimMode:                          simMode,
		FixedIterations:                  fixedIterations,
	}
}

// updateProps holds the properties for the update Amplitude event.
type updateProps struct {
	Success                      bool    `json:"success"`
	MillisForUpdate              int64   `json:"millis_for_update"`
	SecondsSinceLastUpdateMetric float64 `json:"seconds_since_last_update_metric"` // helps understand total runtime of bot when summing this field across events
	NumPruneOps                  int     `json:"num_prune_ops"`
	NumUpdateOpsDelete           int     `json:"num_update_ops_delete"`
	NumUpdateOpsUpdate           int     `json:"num_update_ops_update"`
	NumUpdateOpsCreate           int     `json:"num_update_ops_create"`
}

// deleteProps holds the properties for the delete Amplitude event.
// TODO DS StackTrace may need to be a message instead of or in addition to a
// stack trace. The goal is to get crash logs, Amplitude may not enable this.
type deleteProps struct {
	Exit       bool   `json:"exit"`
	StackTrace string `json:"stack_trace"`
}

type eventWrapper struct {
	APIKey string  `json:"api_key"`
	Events []event `json:"events"`
}

// UpdateLoopResult contains the results of the orderbook update.
// Note that this is used in `trader/trader.go`, but it is defined here to avoid an import cycle.
type UpdateLoopResult struct {
	Success            bool
	NumPruneOps        int
	NumUpdateOpsDelete int
	NumUpdateOpsUpdate int
	NumUpdateOpsCreate int
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

// MakeMetricsTracker is a factory method
func MakeMetricsTracker(
	client *http.Client,
	apiKey string,
	userID string,
	guiUserID string,
	deviceID string,
	botStartTime time.Time,
	isDisabled bool,
	commonProps *CommonPropsStruct,
	cliProps *CliPropsStruct,
) (*MetricsTracker, error) {
	if commonProps == nil {
		return nil, fmt.Errorf("need a non-nil commonProps to make a metrics tracker")
	}

	commonPropsMap, e := utils.ToMapStringInterface(*commonProps)
	if e != nil {
		return nil, fmt.Errorf("could not convert commonProps to map: %s", e)
	}

	var cliPropsMap map[string]interface{}
	if cliProps != nil {
		cliPropsMap, e = utils.ToMapStringInterface(*cliProps)
		if e != nil {
			return nil, fmt.Errorf("could not convert cliProps to map: %s", e)
		}
	}

	mergedProps, e := utils.MergeMaps(commonPropsMap, cliPropsMap)
	if e != nil {
		return nil, fmt.Errorf("could not merge commonPropsMap and cliPropsMap: %s", e)
	}

	return &MetricsTracker{
		client:              client,
		apiKey:              apiKey,
		userID:              userID,
		guiUserID:           guiUserID,
		deviceID:            deviceID,
		eventProps:          mergedProps,
		botStartTime:        botStartTime,
		isDisabled:          isDisabled,
		updateEventSentTime: nil,
		cliVersion:          commonProps.CliVersion,
		failedStartupSend:   false,
	}, nil
}

// GetUpdateEventSentTime gets the last sent time of the update event.
func (mt *MetricsTracker) GetUpdateEventSentTime() *time.Time {
	return mt.updateEventSentTime
}

// SendStartupEvent sends the startup Amplitude event.
func (mt *MetricsTracker) SendStartupEvent(now time.Time) error {
	e := mt.sendEvent(startupEventName, mt.eventProps, now)
	if e != nil {
		mt.failedStartupSend = true
		return fmt.Errorf("metric - failed to send startup event: %s", e)
	}
	return nil
}

// SendUpdateEvent sends the update Amplitude event.
func (mt *MetricsTracker) SendUpdateEvent(now time.Time, updateResult UpdateLoopResult, millisForUpdate int64) error {
	var secondsSinceLastUpdateMetric float64
	if mt.updateEventSentTime == nil {
		secondsSinceLastUpdateMetric = now.Sub(mt.botStartTime).Seconds()
	} else {
		secondsSinceLastUpdateMetric = now.Sub(*mt.updateEventSentTime).Seconds()
	}

	updateProps := updateProps{
		Success:                      updateResult.Success,
		MillisForUpdate:              millisForUpdate,
		SecondsSinceLastUpdateMetric: secondsSinceLastUpdateMetric,
		NumPruneOps:                  updateResult.NumPruneOps,
		NumUpdateOpsDelete:           updateResult.NumUpdateOpsDelete,
		NumUpdateOpsUpdate:           updateResult.NumUpdateOpsUpdate,
		NumUpdateOpsCreate:           updateResult.NumUpdateOpsCreate,
	}

	e := mt.sendEvent(updateEventName, updateProps, now)
	if e != nil {
		return fmt.Errorf("could not send update event: %s", e)
	}

	mt.updateEventSentTime = &now
	return nil
}

// SendDeleteEvent sends the delete Amplitude event.
func (mt *MetricsTracker) SendDeleteEvent(exit bool) error {
	deleteProps := deleteProps{
		Exit:       exit,
		StackTrace: string(debug.Stack()),
	}

	return mt.sendEvent(deleteEventName, deleteProps, time.Now())
}

// SendEvent sends an event with its type and properties to Amplitude.
func (mt *MetricsTracker) sendEvent(eventType string, eventPropsInterface interface{}, now time.Time) error {
	return mt.SendEventForGuiUser(mt.guiUserID, eventType, eventPropsInterface, now)
}

// SendEventForGuiUser sends an event with its type and properties to Amplitude.
func (mt *MetricsTracker) SendEventForGuiUser(guiUserID string, eventType string, eventPropsInterface interface{}, now time.Time) error {
	if mt == nil || mt.apiKey == "" || mt.userID == "-1" || mt.isDisabled {
		log.Printf("metric - not sending event metric of type '%s' because metrics are disabled", eventType)
		return nil
	}

	if mt.failedStartupSend {
		return fmt.Errorf("metric - not sending event metric of type '%s' because we failed to send startup event", eventType)
	}

	trackerProps := mt.eventProps
	trackerProps[secondsSinceStartKey] = now.Sub(mt.botStartTime).Seconds()
	if guiUserID != "" {
		// add this on for requests that come from the gui either as a web event or a bot spawned from the GUI
		trackerProps[guiUserIDKey] = guiUserID
	}

	inputProps, e := utils.ToMapStringInterface(eventPropsInterface)
	if e != nil {
		return fmt.Errorf("could not convert event inputProps to map: %s", e)
	}

	mergedProps, e := utils.MergeMaps(trackerProps, inputProps)
	if e != nil {
		return fmt.Errorf("could not merge event properties: %s", e)
	}

	// session_id is the start time of the session in milliseconds since epoch (Unix Timestamp),
	// necessary to associate events with a particular system (taken from amplitude docs)
	eventW := eventWrapper{
		APIKey: mt.apiKey,
		Events: []event{{
			UserID:          mt.userID,
			SessionID:       mt.botStartTime.Unix() * 1000, // convert to millis based on docs
			DeviceID:        mt.deviceID,
			EventType:       eventType,
			EventProperties: mergedProps,
			Version:         mt.cliVersion,
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
		eventWCensored.APIKey = ""
		requestWCensored, e := json.Marshal(eventWCensored)
		if e != nil {
			return fmt.Errorf("metric - failed to send event metric of type '%s' (response=%s), error while trying to marshall requestWCensored: %s", eventType, responseData.String(), e)
		}
		return fmt.Errorf("metric - failed to send event metric of type '%s' (requestWCensored=%s; response=%s)", eventType, string(requestWCensored), responseData.String())
	}
	return nil
}
