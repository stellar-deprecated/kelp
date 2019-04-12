package utils

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// CheckConfigError checks configs for errors, crashes app if there's an error
func CheckConfigError(cfg fmt.Stringer, e error, filename string) {
	if e != nil {
		log.Println(e)
		log.Println()
		log.Fatalf("error: could not parse the config file '%s'. Check that the correct type of file was passed in.\n", filename)
	}
}

// LogConfig logs out the config file
func LogConfig(cfg fmt.Stringer) {
	log.Println("configs:")
	for _, line := range strings.Split(strings.TrimSuffix(cfg.String(), "\n"), "\n") {
		log.Printf("     %s", line)
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
		isDeprecated := field.Tag.Get("deprecated") == "true"

		// set the transformation function
		transformFn := passthrough
		if fn, ok := transforms[fieldDisplayName]; ok {
			transformFn = fn
		}

		if reflect.ValueOf(s).Field(i).CanInterface() {
			if !isDeprecated || !reflect.ValueOf(s).Field(i).IsNil() {
				value := reflect.ValueOf(s).Field(i).Interface()
				transformedValue := transformFn(value)
				deprecatedWarning := ""
				if isDeprecated {
					deprecatedWarning = " (deprecated)"
				}
				buf.WriteString(fmt.Sprintf("%s: %+v%s\n", fieldDisplayName, transformedValue, deprecatedWarning))
			}
		}
	}
	return buf.String()
}

// SecretKey2PublicKey converts a secret key to a public key
func SecretKey2PublicKey(i interface{}) interface{} {
	if i == "" {
		return ""
	}

	secret, ok := i.(string)
	if !ok {
		log.Fatal("field was not a string")
	}

	pk, e := ParseSecret(secret)
	if e != nil {
		log.Fatal(e)
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

// UnwrapFloat64Pointer unwraps a float64 pointer
func UnwrapFloat64Pointer(i interface{}) interface{} {
	p := i.(*float64)
	if p == nil {
		return ""
	}
	return *p
}

// UnwrapInt8Pointer unwraps a int8 pointer
func UnwrapInt8Pointer(i interface{}) interface{} {
	p := i.(*int8)
	if p == nil {
		return ""
	}
	return *p
}
