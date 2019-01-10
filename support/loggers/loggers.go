package loggers // I didn't call the packacke "logging" because the go-loggeing framework's package is "logging" so we'll need to be different when we switch

import (
	"log"
)

type Logger interface {
	// basic messages, appends a newline (\n) after each entry
	Info(msg string)

	// basic messages, can be custom formatted, similar to fmt.Printf. User needs to add a \n if they want a newline after the log entry
	Infof(msg string, args ...interface{})

	// error messages, indicates to the logger that these messages can be handled differently (different color, special format, email alerts, etc.). The type of logger will determine what to do with these messages. The logger should NOT panic on these messages. Appends a newline (\n) after each entry.
	Error(msg string)

	// error messages, indicates to the logger that these messages can be handled differently (different color, special format, email alerts, etc.). The type of logger will determine what to do with these messages. The logger should NOT panic on these messages. User needs to add a \n if they want a newline after the log entry.
	Errorf(msg string, args ...interface{})

	// added Fatal and Fatalf because trade.go (and elsewhere) use log.Fatal
	// fatal error messages, with newline
	Fatal(e error)

	// formatted fatal error messages, without automatic newline
	Fatalf(e error, args ...interface{})
}

type BasicLogger struct {
}

func (l BasicLogger) Info(msg string) {
	log.Println(msg)
}

func (l BasicLogger) Infof(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (l BasicLogger) Error(msg string) {
	log.Println(msg) // isn't actually differnt from Info until we do structured logs
}

func (l BasicLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...) // isn't actually differnt from Infof until we do structured logs
}

func (l BasicLogger) Fatal(e error) {
	log.Fatal(e)
}

func (l BasicLogger) Fatalf(e string, args ...interface{}) {
	log.Fatalf(e, args...)
}
