package logger

import "log"

// BasicLogger is a standard logger
type BasicLogger struct {
}

// Info impl
func (l *BasicLogger) Info(msg string) {
	log.Println(msg)
}

// Infof impl
func (l *BasicLogger) Infof(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

// Error impl
func (l *BasicLogger) Error(msg string) {
	log.Print(msg)
}

// Efforf impl
func (l *BasicLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg)
}

// ensure it implements Logger
var _ Logger = &BasicLogger{}

// MakeBasicLogger is the factory method
func MakeBasicLogger() Logger {
	var l *BasicLogger
	return l
}
