package plugins

import (
	"log"
	"math/rand"
	"time"

	"github.com/stellar/kelp/api"
)

// IntervalTimeController provides a standard time interval
type IntervalTimeController struct {
	tickInterval time.Duration
	tickDelayFn  func() time.Duration
}

// MakeIntervalTimeController is a factory method
func MakeIntervalTimeController(tickInterval time.Duration, maxTickDelayMillis int64) api.TimeController {
	tickDelayFn := func() time.Duration {
		return time.Duration(0) * time.Millisecond
	}
	if maxTickDelayMillis > 0 {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		tickDelayFn = makeRandomDelayMillisFn(maxTickDelayMillis, randGen)
	}

	return &IntervalTimeController{
		tickInterval: tickInterval,
		tickDelayFn:  tickDelayFn,
	}
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
func (t *IntervalTimeController) SleepTime(lastUpdateTime time.Time) time.Duration {
	// use real time now because we want the start of the clock cycle to be synchronized
	return t.sleepTimeInternal(lastUpdateTime, time.Now())
}

// realNow is the actual current time and not the synchronized time since we want to check sleep from when this function is called
func (t *IntervalTimeController) sleepTimeInternal(lastUpdateTime time.Time, realNow time.Time) time.Duration {
	elapsedSinceUpdate := realNow.Sub(lastUpdateTime)
	fixedDurationCatchup := time.Duration(t.tickInterval.Nanoseconds() - elapsedSinceUpdate.Nanoseconds())
	randomDelayMillis := t.tickDelayFn()

	// if fixedDurationCatchup < 0 then we already have a built-in randomized delay because of the variable processing time consumed
	return fixedDurationCatchup + randomDelayMillis
}

func makeRandomDelayMillisFn(maxTickDelayMillis int64, randGen *rand.Rand) func() time.Duration {
	return func() time.Duration {
		return time.Duration(randGen.Int63n(maxTickDelayMillis)) * time.Millisecond
	}
}
