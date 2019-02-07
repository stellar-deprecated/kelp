package monitoring

import (
	"fmt"
	"log"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/stellar/kelp/api"
)

type pagerDuty struct {
	serviceKey string
}

// ensure pagerDuty implements the api.Alert interface
var _ api.Alert = &pagerDuty{}

func makePagerDuty(serviceKey string) (api.Alert, error) {
	return &pagerDuty{
		serviceKey: serviceKey,
	}, nil
}

// Trigger creates a PagerDuty trigger. The description is required and cannot be empty. Supplementary
// details can be optionally provided as key-value pairs as part of the details parameter.
func (p *pagerDuty) Trigger(description string, details interface{}) error {
	event := pagerduty.Event{
		ServiceKey:  p.serviceKey,
		Type:        "trigger",
		Description: description,
		Details:     details,
	}
	response, e := pagerduty.CreateEvent(event)
	if e != nil {
		return fmt.Errorf("encountered an error while sending a PagerDuty alert: %s", e)
	}
	log.Printf("Triggered PagerDuty alert. Incident key for reference: %s\n", response.IncidentKey)
	return nil
}
