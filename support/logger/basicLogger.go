package logger

import "log"

// basicLogger is a standard logger
type basicLogger struct {
}

// Info impl
func (l *basicLogger) Info(msg string) {
	log.Println(msg)
}

// Infof impl
func (l *basicLogger) Infof(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

// Error impl
func (l *basicLogger) Error(msg string) {
	log.Print(msg)
}

// Efforf impl
func (l *basicLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

// ensure it implements Logger
var _ Logger = &basicLogger{}

// MakeBasicLogger is the factory method
func MakeBasicLogger() Logger {
	return &basicLogger{}
}
