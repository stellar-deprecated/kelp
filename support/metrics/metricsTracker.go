package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/stellar/kelp/support/logger"

	"github.com/google/uuid"
)

var (
	amplitudeAPIURL string = "https://api2.amplitude.com/2/httpapi"
	amplitudeAPIKey string = os.Getenv("AMPLITUDE_API_KEY")
)

// Tracker wraps the properties for Amplitude events,
// and can be used to directly send events to the
// Amplitude HTTP API.
type Tracker struct {
	client *http.Client
	props  props
	start  time.Time
}

type event struct {
	UserID    string      `json:"user_id"`
	DeviceID  string      `json:"device_id"`
	EventType string      `json:"event_type"`
	Props     interface{} `json:"event_properties"`
}

// props holds the properties that we need for all Amplitude events.
// This lives on the `Tracker` struct, although
// TODO: Add geodata.
// TODO: Add cloud server information.
type props struct {
	BuildVersion string    `json:"build_version"`
	Os           string    `json:"os"`
	Gui          bool      `json:"gui"`
	Strategy     string    `json:"strategy"`
	UpdateTime   float64   `json:"update_time"`
	Exchange     string    `json:"exchange"`
	Pair         string    `json:"pair"`
	SessionID    uuid.UUID `json:"session_id"`
}

// updateProps holds the properties for the update Amplitude event.
type updateProps struct {
	props
	SecondsSinceStart float64 `json:"seconds_since_start"`
}

// deleteProps holds the properties for the delete Amplitude event.
type deleteProps struct {
	props
	SecondsSinceStart float64 `json:"seconds_since_start"`
	Exit              bool    `json:"exit"`
	StackTrace        string  `json:"stack_trace"`
}

// MakeMetricsTracker is a factory method to create a `metrics.Tracker`.
func MakeMetricsTracker(
	version string,
	os string,
	gui bool,
	strategy string,
	updateTime float64,
	exchange string,
	pair string,
) *Tracker {
	client := &http.Client{}
	sessionID := mustSessionID()
	props := props{
		BuildVersion: version,
		Os:           os,
		Gui:          gui,
		Strategy:     strategy,
		UpdateTime:   updateTime,
		Exchange:     exchange,
		Pair:         pair,
		SessionID:    sessionID,
	}

	t := Tracker{
		client: client,
		props:  props,
		start:  time.Now(),
	}
	return &t
}

func mustSessionID() uuid.UUID {
	sessionID, e := uuid.NewRandom()
	if e != nil {
		return [16]byte{}
	}

	return sessionID
}

// SendStartupEvent sends the startup Amplitude event.
func (t *Tracker) SendStartupEvent(l logger.Logger) error {
	return t.sendEvent(l, "ce:test_startup", t.props)
}

// SendUpdateEvent sends the update Amplitude event.
func (t *Tracker) SendUpdateEvent(l logger.Logger) error {
	props := updateProps{
		props:             t.props,
		SecondsSinceStart: time.Now().Sub(t.start).Seconds(),
	}

	return t.sendEvent(l, "ce:test_update", props)
}

// SendDeleteEvent sends the delete Amplitude event.
func (t *Tracker) SendDeleteEvent(l logger.Logger, exit bool) error {
	props := deleteProps{
		props:             t.props,
		SecondsSinceStart: time.Now().Sub(t.start).Seconds(),
		Exit:              exit,
		StackTrace:        "", // TODO: Determine how to do this.
	}

	return t.sendEvent(l, "ce:test_delete", props)
}

func (t *Tracker) sendEvent(l logger.Logger, eventType string, eventProps interface{}) error {
	requestBody, e := json.Marshal(map[string]interface{}{
		"api_key": amplitudeAPIKey,
		"events": []event{event{
			UserID:    "12345", // TODO: Determine actual user id.
			EventType: eventType,
			Props:     eventProps,
		}},
	})

	if e != nil {
		l.Info("could not marshal json request")
		return e
	}

	resp, e := http.Post(amplitudeAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if e != nil {
		l.Info("could not post amplitude request")
		return e
	}

	defer resp.Body.Close()

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		l.Info("could not read response body")
		return e
	}

	l.Info(fmt.Sprintf("Successfully sent startup event with response %v\n", string(body)))
	return nil
}
