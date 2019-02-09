package plugins

import (
	"math/rand"
	"time"

	"github.com/stellar/kelp/support/logger"

	"github.com/stellar/kelp/api"
)

// IntervalTimeController provides a standard time interval
type IntervalTimeController struct {
	tickInterval       time.Duration
	maxTickDelayMillis int64
	randGen            *rand.Rand
	l                  logger.Logger
}

// MakeIntervalTimeController is a factory method
func MakeIntervalTimeController(tickInterval time.Duration, maxTickDelayMillis int64, l logger.Logger) api.TimeController {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &IntervalTimeController{
		tickInterval:       tickInterval,
		maxTickDelayMillis: maxTickDelayMillis,
		randGen:            randGen,
		l:                  l,
	}
}

var _ api.TimeController = &IntervalTimeController{}

// ShouldUpdate impl
func (t *IntervalTimeController) ShouldUpdate(lastUpdateTime time.Time, currentUpdateTime time.Time) bool {
	elapsedSinceUpdate := currentUpdateTime.Sub(lastUpdateTime)
	shouldUpdate := elapsedSinceUpdate >= t.tickInterval
	t.l.Infof("intervalTimeController tickInterval=%s, shouldUpdate=%v, elapsedSinceUpdate=%s\n", t.tickInterval, shouldUpdate, elapsedSinceUpdate)
	return shouldUpdate
}

// SleepTime impl
func (t *IntervalTimeController) SleepTime(lastUpdateTime time.Time, currentUpdateTime time.Time) time.Duration {
	// use time till now as opposed to currentUpdateTime because we want the start of the clock cycle to be synchronized
	elapsedSinceUpdate := time.Since(lastUpdateTime)
	fixedDurationCatchup := time.Duration(t.tickInterval.Nanoseconds() - elapsedSinceUpdate.Nanoseconds())
	randomizedDelayMillis := t.makeRandomDelay()

	// if fixedDurationCatchup < 0 then we already have a built-in randomized delay because of the variable processing time consumed
	return fixedDurationCatchup + randomizedDelayMillis
}

func (t *IntervalTimeController) makeRandomDelay() time.Duration {
	if t.maxTickDelayMillis > 0 {
		return time.Duration(t.randGen.Int63n(t.maxTickDelayMillis)) * time.Millisecond
	}
	return time.Duration(0) * time.Millisecond
}
