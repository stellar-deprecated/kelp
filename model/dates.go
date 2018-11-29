package model

import (
	"fmt"
	"time"
)

// Timestamp is millis since epoch
type Timestamp int64

// MakeTimestamp creates a new Timestamp
func MakeTimestamp(ts int64) *Timestamp {
	timestamp := Timestamp(ts)
	return &timestamp
}

// MakeTimestampFromTime creates a new Timestamp
func MakeTimestampFromTime(t time.Time) *Timestamp {
	return MakeTimestamp(t.UnixNano() / int64(time.Millisecond))
}

func (t *Timestamp) String() string {
	return fmt.Sprintf("%d", t.AsInt64())
}

// AsInt64 is a convenience method
func (t *Timestamp) AsInt64() int64 {
	return int64(*t)
}
