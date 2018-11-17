package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggerPagerDuty(t *testing.T) {
	if testing.Short() {
		return
	}

	const kelpServiceKey = "" // Fill in pager duty service key here during testing.
	testCases := []struct {
		testName      string
		serviceKey    string
		description   string
		details       interface{}
		errorExpected bool
	}{
		{
			testName:      "Tests that a Pager Duty alert was triggered successfully",
			serviceKey:    kelpServiceKey,
			description:   "Testing monitoring package. Not a real incident!",
			details:       nil,
			errorExpected: false,
		},
		// Description cannot be empty
		{
			testName:      "Tests that a missing description causes an error",
			serviceKey:    kelpServiceKey,
			description:   "",
			details:       nil,
			errorExpected: true,
		},
		// Service key is invalid
		{
			testName:      "Tests that an invalid API key causes an error",
			serviceKey:    "",
			description:   "Testing monitoring package. Not a real incident!",
			details:       nil,
			errorExpected: true,
		},
		{
			testName:    "Tests that details can be passed through",
			serviceKey:  kelpServiceKey,
			description: "Testing monitoring package. Not a real incident!",
			details: struct {
				LoadAvg     float64 `json:"load_avg"`
				NumRequests int     `json:"num_requests"`
			}{
				LoadAvg:     0.5,
				NumRequests: 100,
			},
			errorExpected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			pagerDutyAlert, e := MakeAlert("PagerDuty", tc.serviceKey)
			if !assert.Nil(t, e) {
				return
			}
			e = pagerDutyAlert.Trigger(tc.description, tc.details)
			if tc.errorExpected {
				assert.NotNil(t, e)
			} else {
				assert.Nil(t, e)
			}
		})
	}
}
