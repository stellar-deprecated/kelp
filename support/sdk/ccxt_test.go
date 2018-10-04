package sdk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeValid(t *testing.T) {
	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "kraken")
	if e != nil {
		assert.Fail(t, fmt.Sprintf("unexpected error: %s", e))
		return
	}
	// success
}

func TestMakeInvalid(t *testing.T) {
	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "missing-exchange")
	if e == nil {
		assert.Fail(t, "expected an error when trying to make and initialize an exchange that is missing: 'missing-exchange'")
		return
	}

	if !strings.Contains(e.Error(), "exchange name 'missing-exchange' is not in the list") {
		assert.Fail(t, fmt.Sprintf("unexpected error: %s", e))
		return
	}
	// success
}
