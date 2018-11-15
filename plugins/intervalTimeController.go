package plugins

import (
	"log"
	"time"

	"github.com/lightyeario/kelp/api"
)

// IntervalTimeController provides a standard time interval
type IntervalTimeController struct {
	tickInterval time.Duration
}

// MakeIntervalTimeController is a factory method
func MakeIntervalTimeController(tickInterval time.Duration) api.TimeController {
	return &IntervalTimeController{tickInterval: tickInterval}
}

var _ api.TimeController = &IntervalTimeController{}

// ShouldUpdate impl
func (t *IntervalTimeController) ShouldUpdate(lastUpdateTime time.Time, currentUpdateTime time.Time) bool {
	elapsedSinceUpdate := currentUpdateTime.Sub(lastUpdateTime)
	shouldUpdate := elapsedSinceUpdate >= t.tickInterval
	log.Printf("intervalTimeController tickInterval=%s, shouldUpdate=%v, elapsedSinceUpdate=%s\n", t.tickInterval, shouldUpdate, elapsedSinceUpdate)
	return shouldUpdate
}

// SleepTime impl
func (t *IntervalTimeController) SleepTime(lastUpdateTime time.Time, currentUpdateTime time.Time) time.Duration {
	// use time till now as opposed to currentUpdateTime because we want the start of the clock cycle to be synchronized
	elapsedSinceUpdate := time.Since(lastUpdateTime)
	return time.Duration(t.tickInterval.Nanoseconds() - elapsedSinceUpdate.Nanoseconds())
}
