package api

import "time"

// TimeController controls the update loop for the bot
type TimeController interface {
	// ShouldUpdate defines when to enter the bot's update cycle
	// lastUpdateTime will never start off as the zero value
	ShouldUpdate(lastUpdateTime time.Time, currentUpdateTime time.Time) bool

	// SleepTime computes how long we want to sleep before the next call to ShouldUpdate
	SleepTime(lastUpdateTime time.Time) time.Duration
}
