package utils

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/interstellar/kelp/support/logger"
)

// CheckConfigError checks configs for errors, crashes app if there's an error
func CheckConfigError(l logger.Logger, cfg fmt.Stringer, e error, filename string) {
	if e != nil {
		logger.Fatal(utilsLogger, fmt.Errorf("error: could not parse the config file '%s'. Check that the correct type of file was passed in", filename))
	}
}

// LogConfig logs out the config file
func LogConfig(cfg fmt.Stringer) {
	utilsLogger.Info("configs:")
	for _, line := range strings.Split(strings.TrimSuffix(cfg.String(), "\n"), "\n") {
		utilsLogger.Infof("     %s", line)
	}
}

// StructString is a helper method that
func StructString(s interface{}, transforms map[string]func(interface{}) interface{}) string {
	var buf bytes.Buffer
	numFields := reflect.TypeOf(s).NumField()
	for i := 0; i < numFields; i++ {
		field := reflect.TypeOf(s).Field(i)
		fieldName := field.Name
		fieldDisplayName := field.Tag.Get("toml")
		if fieldDisplayName == "" {
			fieldDisplayName = fieldName
		}

		// set the transformation function
		transformFn := passthrough
		if fn, ok := transforms[fieldDisplayName]; ok {
			transformFn = fn
		}

		if reflect.ValueOf(s).Field(i).CanInterface() {
			value := reflect.ValueOf(s).Field(i).Interface()
			transformedValue := transformFn(value)
			buf.WriteString(fmt.Sprintf("%s: %+v\n", fieldDisplayName, transformedValue))
		}
	}
	return buf.String()
}

// SecretKey2PublicKey converts a secret key to a public key
func SecretKey2PublicKey(l logger.Logger, i interface{}) interface{} {
	if i == "" {
		return ""
	}

	secret, ok := i.(string)
	if !ok {
		logger.Fatal(utilsLogger, fmt.Errorf("field was not a string"))
	}

	pk, e := ParseSecret(secret)
	if e != nil {
		logger.Fatal(utilsLogger, e)
	}
	return fmt.Sprintf("[secret key to account %s]", *pk)
}

// Passthrough returns the input
func passthrough(i interface{}) interface{} {
	return i
}

// Hide returns an empty string
func Hide(i interface{}) interface{} {
	return ""
}
