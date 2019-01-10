package logger

import "log"

<<<<<<< HEAD
// basicLogger is a standard logger
type basicLogger struct {
}

// Info impl
func (l *basicLogger) Info(msg string) {
=======
// BasicLogger is a standard logger
type BasicLogger struct {
}

// Info impl
func (l *BasicLogger) Info(msg string) {
>>>>>>> major rework
	log.Println(msg)
}

// Infof impl
<<<<<<< HEAD
func (l *basicLogger) Infof(msg string, args ...interface{}) {
=======
func (l *BasicLogger) Infof(msg string, args ...interface{}) {
>>>>>>> major rework
	log.Printf(msg, args...)
}

// Error impl
<<<<<<< HEAD
func (l *basicLogger) Error(msg string) {
=======
func (l *BasicLogger) Error(msg string) {
>>>>>>> major rework
	log.Print(msg)
}

// Efforf impl
<<<<<<< HEAD
func (l *basicLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

// ensure it implements Logger
var _ Logger = &basicLogger{}

// MakeBasicLogger is the factory method
func MakeBasicLogger() Logger {
	return &basicLogger{}
=======
func (l *BasicLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg)
}

// ensure it implements Logger
var _ Logger = &BasicLogger{}

// MakeBasicLogger is the factory method
func MakeBasicLogger() Logger {
	var l *BasicLogger
	return l
>>>>>>> major rework
}
