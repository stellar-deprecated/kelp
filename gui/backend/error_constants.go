package backend

// errorType is a constant string for error types
type errorType string

// enum values
const (
	errorTypeBot errorType = "object_type_bot"
)

// String is the Stringer method
func (et errorType) String() string {
	return string(et)
}

// errorLevel is a constant string for error levels
type errorLevel string

// enum values
const (
	errorLevelInfo    errorLevel = "info"
	errorLevelWarning errorLevel = "warning"
	errorLevelError   errorLevel = "error"
)

// String is the Stringer method
func (el errorLevel) String() string {
	return string(el)
}
