package plugins

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldUpdate(t *testing.T) {
	type testStruct struct {
		name                  string
		tickInterval          time.Duration
		delayMillis           int64
		millisSinceLastUpdate int64
		wantShouldUpdate      bool
	}
	testCases := []testStruct{}

	// random delay has no effect on the output; test different values to ensure that
	for _, delayMillis := range []int64{0, 100, 1000, 10000} {
		testCases = append(testCases, []testStruct{
			{
				name:                  "no change",
				tickInterval:          time.Duration(5 * time.Second),
				delayMillis:           delayMillis,
				millisSinceLastUpdate: 0,
				wantShouldUpdate:      false,
			}, {
				name:                  "inside",
				tickInterval:          time.Duration(5 * time.Second),
				delayMillis:           delayMillis,
				millisSinceLastUpdate: 4999,
				wantShouldUpdate:      false,
			}, {
				name:                  "exact",
				tickInterval:          time.Duration(5 * time.Second),
				delayMillis:           delayMillis,
				millisSinceLastUpdate: 5000,
				wantShouldUpdate:      true,
			}, {
				name:                  "outside",
				tickInterval:          time.Duration(5 * time.Second),
				delayMillis:           delayMillis,
				millisSinceLastUpdate: 5001,
				wantShouldUpdate:      true,
			}, {
				name:                  "very much outside",
				tickInterval:          time.Duration(5 * time.Second),
				delayMillis:           delayMillis,
				millisSinceLastUpdate: 50000,
				wantShouldUpdate:      true,
			},
		}...)
	}

	for _, k := range testCases {
		t.Run(fmt.Sprintf("%s, delayMillis=%d", k.name, k.delayMillis), func(t *testing.T) {
			delay := time.Millisecond * time.Duration(k.delayMillis)
			tc := makeIntervalTimeControllerForTest(k.tickInterval, delay)

			lastUpdateTime, _ := time.Parse(time.RFC3339, "2020-03-14T15:00:00Z")
			currentUpdateTime := lastUpdateTime.Add(time.Millisecond * time.Duration(k.millisSinceLastUpdate))
			gotShouldUpdate := tc.ShouldUpdate(lastUpdateTime, currentUpdateTime)

			assert.Equal(t, k.wantShouldUpdate, gotShouldUpdate)
		})
	}
}

func TestSleepTime(t *testing.T) {
	testCases := []struct {
		name                  string
		delayMillis           int64
		millisSinceLastUpdate int64
		wantDuration          time.Duration
	}{
		{
			name:                  "no delay, no time diff",
			delayMillis:           0,
			millisSinceLastUpdate: 0,
			wantDuration:          time.Duration(5000) * time.Millisecond,
		}, {
			name:                  "no delay, 1 ms elapsed",
			delayMillis:           0,
			millisSinceLastUpdate: 1,
			wantDuration:          time.Duration(4999) * time.Millisecond,
		}, {
			name:                  "no delay, 4999 ms elapsed",
			delayMillis:           0,
			millisSinceLastUpdate: 4999,
			wantDuration:          time.Duration(1) * time.Millisecond,
		}, {
			name:                  "no delay, exact time update",
			delayMillis:           0,
			millisSinceLastUpdate: 5000,
			wantDuration:          time.Duration(0) * time.Millisecond,
		}, {
			name:                  "no delay, surpassed one update cycle",
			delayMillis:           0,
			millisSinceLastUpdate: 5001,
			wantDuration:          time.Duration(-1) * time.Millisecond,
		}, {
			name:                  "no delay, surpassed one update cycle by a lot",
			delayMillis:           0,
			millisSinceLastUpdate: 15001,
			wantDuration:          time.Duration(-10001) * time.Millisecond,
		}, {
			name:                  "delay, no time diff",
			delayMillis:           1,
			millisSinceLastUpdate: 0,
			wantDuration:          time.Duration(5001) * time.Millisecond,
		}, {
			name:                  "delay, 4999 ms elapsed",
			delayMillis:           1023,
			millisSinceLastUpdate: 4999,
			wantDuration:          time.Duration(1024) * time.Millisecond,
		}, {
			name:                  "delay, exact time update",
			delayMillis:           1023,
			millisSinceLastUpdate: 6023,
			wantDuration:          time.Duration(0) * time.Millisecond,
		},
	}

	for i, k := range testCases {
		name := fmt.Sprintf("%d. %s", (i + 1), k.name)
		t.Run(name, func(t *testing.T) {
			delay := time.Millisecond * time.Duration(k.delayMillis)
			tc := makeIntervalTimeControllerForTest(time.Duration(5*time.Second), delay)

			lastUpdateTime, _ := time.Parse(time.RFC3339, "2020-03-14T15:00:00Z")
			realNow := lastUpdateTime.Add(time.Millisecond * time.Duration(k.millisSinceLastUpdate))
			gotDuration := tc.sleepTimeInternal(lastUpdateTime, realNow)

			assert.Equal(t, k.wantDuration, gotDuration)
		})
	}
}

// factory method takes a deterministic delay for tests
func makeIntervalTimeControllerForTest(tickInterval time.Duration, delay time.Duration) *IntervalTimeController {
	tickDelayFn := func() time.Duration {
		return delay
	}
	return &IntervalTimeController{
		tickInterval: tickInterval,
		tickDelayFn:  tickDelayFn,
	}
}
