package logger

import "log"

<<<<<<< HEAD
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
=======
// basicLogger is a standard logger
type basicLogger struct {
}

// Info impl
func (l *basicLogger) Info(msg string) {
>>>>>>> remove gloval var, format edits
	log.Println(msg)
}

// Infof impl
<<<<<<< HEAD
<<<<<<< HEAD
func (l *basicLogger) Infof(msg string, args ...interface{}) {
=======
func (l *BasicLogger) Infof(msg string, args ...interface{}) {
>>>>>>> major rework
=======
func (l *basicLogger) Infof(msg string, args ...interface{}) {
>>>>>>> remove gloval var, format edits
	log.Printf(msg, args...)
}

// Error impl
<<<<<<< HEAD
<<<<<<< HEAD
func (l *basicLogger) Error(msg string) {
=======
func (l *BasicLogger) Error(msg string) {
>>>>>>> major rework
=======
func (l *basicLogger) Error(msg string) {
>>>>>>> remove gloval var, format edits
	log.Print(msg)
}

// Efforf impl
<<<<<<< HEAD
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
=======
func (l *basicLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
>>>>>>> remove gloval var, format edits
}

// ensure it implements Logger
var _ Logger = &basicLogger{}

// MakeBasicLogger is the factory method
func MakeBasicLogger() Logger {
<<<<<<< HEAD
	var l *BasicLogger
	return l
>>>>>>> major rework
=======
	return &basicLogger{}
>>>>>>> remove gloval var, format edits
}
