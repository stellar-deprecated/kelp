package utils

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// CheckConfigError checks configs for errors
func CheckConfigError(cfg fmt.Stringer, e error, filename string) {
	if e != nil {
		log.Println(e)
		log.Println()
		log.Fatalf("error: could not parse the config file '%s'. Check that the correct type of file was passed in.\n", filename)
	}

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
		fieldName := reflect.TypeOf(s).Field(i).Name

		// set the transformation function
		transformFn := passthrough
		if fn, ok := transforms[fieldName]; ok {
			transformFn = fn
		}

		if reflect.ValueOf(s).Field(i).CanInterface() {
			value := reflect.ValueOf(s).Field(i).Interface()
			transformedValue := transformFn(value)
			buf.WriteString(fmt.Sprintf("%s: %+v\n", fieldName, transformedValue))
		}
	}
	return buf.String()
}

// SecretKey2PublicKey converts a secret key to a public key
func SecretKey2PublicKey(i interface{}) interface{} {
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
