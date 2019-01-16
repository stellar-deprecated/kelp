package logger

import "os"

// Logger is the base logger interface
type Logger interface {
	// basic messages, appends a newline (\n) after each entry
	Info(msg string)

	// basic messages, can be custom formatted, similar to fmt.Printf. User needs to add a \n if they want a newline after the log entry
	Infof(msg string, args ...interface{})

	// error messages, indicates to the logger that these messages can be handled differently (different color, special format, email alerts, etc.). The type of logger will determine what to do with these messages. The logger should NOT panic on these messages. Appends a newline (\n) after each entry.
	Error(msg string)

	// error messages, indicates to the logger that these messages can be handled differently (different color, special format, email alerts, etc.). The type of logger will determine what to do with these messages. The logger should NOT panic on these messages. User needs to add a \n if they want a newline after the log entry.
	Errorf(msg string, args ...interface{})
}

// Fatal is a convenience method that can be used with any Logger to log a fatal error
func Fatal(l Logger, e error) {
	l.Info("")
	l.Errorf("%s", e)
	os.Exit(1)
}
