package model

// Timestamp is millis since epoch
type Timestamp int64

// MakeTimestamp creates a new Timestamp
func MakeTimestamp(ts int64) *Timestamp {
	timestamp := Timestamp(ts)
	return &timestamp
}

// AsInt64 is a convenience method
func (t Timestamp) AsInt64() int64 {
	return int64(t)
}
