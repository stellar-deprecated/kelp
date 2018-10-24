package monitoring

import (
	"fmt"

	"github.com/lightyeario/kelp/api"
)

// MakeAlert creates an Alert based on the type of the service (eg Pager Duty) and its corresponding API key.
func MakeAlert(alertType string, apiKey string) (api.Alert, error) {
	switch alertType {
	case "PagerDuty":
		return makePagerDuty(apiKey)
	default:
		return nil, fmt.Errorf("cannot make alert - invalid alert type: %s", alertType)
	}
}
